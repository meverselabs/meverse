package main

import (
	"encoding/hex"
	"io/ioutil"
	"math/big"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/fletaio/fleta_v2/cmd/app"
	"github.com/fletaio/fleta_v2/cmd/closer"
	"github.com/fletaio/fleta_v2/cmd/config"
	"github.com/fletaio/fleta_v2/common"
	"github.com/fletaio/fleta_v2/common/hash"
	"github.com/fletaio/fleta_v2/common/key"
	"github.com/fletaio/fleta_v2/core/chain"
	"github.com/fletaio/fleta_v2/core/piledb"
	"github.com/fletaio/fleta_v2/core/types"
	"github.com/fletaio/fleta_v2/node"
	"github.com/fletaio/fleta_v2/service/account"
	"github.com/fletaio/fleta_v2/service/apiserver"
	"github.com/fletaio/fleta_v2/service/apiserver/metamaskrelay"
	"github.com/fletaio/fleta_v2/service/apiserver/viewchain"
	"github.com/fletaio/fleta_v2/service/txsearch"
)

// Config is a configuration for the cmd
type Config struct {
	SeedNodeMap     map[string]string
	ObserverMap     map[string]string
	GeneratorKeyHex string
	NodeKeyHex      string
	InitGenesisHash string
	InitHash        string
	InitHeight      uint32
	InitTimestamp   uint64
	Port            int
	StoreRoot       string
	UseWSS          bool
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

	var frkey key.Key
	if len(cfg.GeneratorKeyHex) == 0 {
		panic("not exist generator key")
	}
	if bs, err := hex.DecodeString(cfg.GeneratorKeyHex); err != nil {
		panic(err)
	} else if Key, err := key.NewMemoryKeyFromBytes(ChainID, bs); err != nil {
		panic(err)
	} else {
		frkey = Key
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
	ObserverNodeMap := map[common.PublicKey]string{}
	for k, v := range cfg.ObserverMap {
		pubkey, err := common.ParsePublicKey(k)
		if err != nil {
			panic(err)
		}
		ObserverKeys = append(ObserverKeys, pubkey)

		if cfg.UseWSS {
			ObserverNodeMap[pubkey] = "wss://" + v
		} else {
			ObserverNodeMap[pubkey] = "ws://" + v
		}
	}
	SeedNodeMap := map[common.PublicKey]string{}
	for k, netAddr := range cfg.SeedNodeMap {
		pubhash, err := common.ParsePublicKey(k)
		if err != nil {
			panic(err)
		}
		SeedNodeMap[pubhash] = netAddr
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
	api := account.NewAccountAPI()
	rpcapi := apiserver.NewAPIServer()
	ts := txsearch.NewTxSearch(cfg.StoreRoot+"/_txsearch", rpcapi, st, cn)
	cn.MustAddService(ts)
	cn.MustAddService(api)
	cn.MustAddService(rpcapi)
	if cfg.InitHeight == 0 {
		if err := cn.Init(app.Genesis()); err != nil {
			panic(err)
		}
	} else {
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

	fr := node.NewGeneratorNode(ChainID, &node.GeneratorConfig{
		MaxTransactionsPerBlock: 20000,
	}, cn, frkey, ndkey, ObserverNodeMap, SeedNodeMap, cfg.StoreRoot+"/peer")
	if err := fr.Init(); err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("formulator", fr)

	metamaskrelay.NewMetamaskRelay(rpcapi, ts, cn, fr)
	go rpcapi.Run(":8541")
	go api.Run(cn, fr, st)
	viewchain.NewViewchain(rpcapi, ts, cn, st, fr)

	go fr.Run(":" + strconv.Itoa(cfg.Port))

	cm.Wait()
}
