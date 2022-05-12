package main

import (
	"encoding/hex"
	"io/ioutil"
	"math/big"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/meverselabs/meverse/cmd/app"
	"github.com/meverselabs/meverse/cmd/closer"
	"github.com/meverselabs/meverse/cmd/config"
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/piledb"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/p2p"
	"github.com/meverselabs/meverse/service/apiserver"
	"github.com/meverselabs/meverse/service/apiserver/metamaskrelay"
	"github.com/meverselabs/meverse/service/apiserver/viewchain"
	"github.com/meverselabs/meverse/service/txsearch"
)

// Config is a configuration for the cmd
type Config struct {
	SeedNodeMap     map[string]string
	NodeKeyHex      string
	ObserverKeys    []string
	InitGenesisHash string
	InitHash        string
	InitHeight      uint32
	InitTimestamp   uint64
	Port            int
	RPCPort         int
	StoreRoot       string
}

func main() {
	ChainID := big.NewInt(0x1D5E)
	Version := uint16(0x0001)
	var cfg Config
	if err := config.LoadFile("./config.toml", &cfg); err != nil {
		panic(err)
	}
	if len(cfg.StoreRoot) == 0 {
		cfg.StoreRoot = "./ndata"
	}

	var ndkey key.Key
	if len(cfg.NodeKeyHex) > 0 {
		if bs, err := hex.DecodeString(cfg.NodeKeyHex); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(ChainID, bs); err != nil {
			panic(err)
		} else {
			ndkey = Key
		}
	} else {
		if bs, err := ioutil.ReadFile("./ndkey.key"); err != nil {
			k, err := key.NewMemoryKey(ChainID)
			if err != nil {
				panic(err)
			}

			fs, err := os.Create("./ndkey.key")
			if err != nil {
				panic(err)
			}
			fs.Write(k.Bytes())
			fs.Close()
			ndkey = k
		} else {
			if Key, err := key.NewMemoryKeyFromBytes(ChainID, bs); err != nil {
				panic(err)
			} else {
				ndkey = Key
			}
		}
	}

	ObserverKeys := []common.PublicKey{}
	for _, k := range cfg.ObserverKeys {
		pubkey, err := common.ParsePublicKey(k)
		if err != nil {
			panic(err)
		}
		ObserverKeys = append(ObserverKeys, pubkey)
	}
	SeedNodeMap := map[common.PublicKey]string{}
	for k, netAddr := range cfg.SeedNodeMap {
		pubkey, err := common.ParsePublicKey(k)
		if err != nil {
			panic(err)
		}
		SeedNodeMap[pubkey] = netAddr
	}

	cm := closer.NewManager()
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		<-sigc
		cm.CloseAll()
	}()
	defer cm.CloseAll()

	//MaxBlocksPerFormulator := uint32(10)

	var InitGenesisHash hash.Hash256
	if len(cfg.InitGenesisHash) > 0 {
		InitGenesisHash = hash.HexToHash(cfg.InitGenesisHash)
	}
	var InitHash hash.Hash256
	if len(cfg.InitHash) > 0 {
		InitHash = hash.HexToHash(cfg.InitHash)
	}

	cdb, err := piledb.Open(cfg.StoreRoot+"/chain", InitHash, cfg.InitHeight, cfg.InitTimestamp)
	if err != nil {
		panic(err)
	}
	cdb.SetSyncMode(true)
	st, err := chain.NewStore(cfg.StoreRoot+"/context", cdb, ChainID, Version)
	if err != nil {
		panic(err)
	}
	cm.Add("store", st)

	if st.Height() > st.InitHeight() {
		if _, err := cdb.GetData(st.Height(), 0); err != nil {
			panic(err)
		}
	}

	cn := chain.NewChain(ObserverKeys, st, "")
	rpcapi := apiserver.NewAPIServer()
	ts := txsearch.NewTxSearch(cfg.StoreRoot+"/_txsearch", rpcapi, st, cn, cfg.InitHeight)
	cn.MustAddService(ts)
	cn.MustAddService(rpcapi)

	if cfg.InitHeight == 0 {
		if err := cn.Init(app.Genesis()); err != nil {
			panic(err)
		}
	} else {
		app.RegisterContractClass()
		if err := cn.InitWith(InitGenesisHash, InitHash, cfg.InitHeight, cfg.InitTimestamp); err != nil {
			panic(err)
		}
	}
	cm.RemoveAll()
	cm.Add("chain", cn)

	if err := st.IterBlockAfterContext(func(b *types.Block) error {
		if cm.IsClosed() {
			return chain.ErrStoreClosed
		}
		if err := cn.ConnectBlock(b, nil); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if err == chain.ErrStoreClosed {
			return
		}
		panic(err)
	}

	nd := p2p.NewNode(ChainID, ndkey, SeedNodeMap, cn, cfg.StoreRoot+"/peer")
	if err := nd.Init(); err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("node", nd)

	metamaskrelay.NewMetamaskRelay(rpcapi, ts, cn, nd)
	go rpcapi.Run(":" + strconv.Itoa(cfg.RPCPort))
	viewchain.NewViewchain(rpcapi, ts, cn, st, nd)

	go nd.Run(":" + strconv.Itoa(cfg.Port))
	cm.Wait()
}
