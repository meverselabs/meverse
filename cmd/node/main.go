package main

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/fletaio/fleta/core/pile"

	"github.com/fletaio/fleta/cmd/app"
	"github.com/fletaio/fleta/cmd/closer"
	"github.com/fletaio/fleta/cmd/config"
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/common/rlog"
	"github.com/fletaio/fleta/core/backend"
	_ "github.com/fletaio/fleta/core/backend/badger_driver"
	_ "github.com/fletaio/fleta/core/backend/buntdb_driver"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/pof"
	"github.com/fletaio/fleta/process/admin"
	"github.com/fletaio/fleta/process/formulator"
	"github.com/fletaio/fleta/process/gateway"
	"github.com/fletaio/fleta/process/payment"
	"github.com/fletaio/fleta/process/vault"
	"github.com/fletaio/fleta/service/apiserver"
	"github.com/fletaio/fleta/service/p2p"
)

// Config is a configuration for the cmd
type Config struct {
	SeedNodeMap    map[string]string
	NodeKeyHex     string
	ObserverKeys   []string
	Port           int
	APIPort        int
	StoreRoot      string
	BackendVersion int
	RLogHost       string
	RLogPath       string
	UseRLog        bool
}

func main() {
	var cfg Config
	if err := config.LoadFile("./config.toml", &cfg); err != nil {
		panic(err)
	}
	if len(cfg.StoreRoot) == 0 {
		cfg.StoreRoot = "./ndata"
	}
	if len(cfg.RLogHost) > 0 && cfg.UseRLog {
		if len(cfg.RLogPath) == 0 {
			cfg.RLogPath = "./ndata_rlog"
		}
		rlog.SetRLogHost(cfg.RLogHost)
		rlog.Enablelogger(cfg.RLogPath)
	}

	var ndkey key.Key
	if len(cfg.NodeKeyHex) > 0 {
		if bs, err := hex.DecodeString(cfg.NodeKeyHex); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
			panic(err)
		} else {
			ndkey = Key
		}
	} else {
		if bs, err := ioutil.ReadFile("./ndkey.key"); err != nil {
			k, err := key.NewMemoryKey()
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
			if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
				panic(err)
			} else {
				ndkey = Key
			}
		}
	}

	ObserverKeys := []common.PublicHash{}
	for _, k := range cfg.ObserverKeys {
		pubhash, err := common.ParsePublicHash(k)
		if err != nil {
			panic(err)
		}
		ObserverKeys = append(ObserverKeys, pubhash)
	}
	SeedNodeMap := map[common.PublicHash]string{}
	for k, netAddr := range cfg.SeedNodeMap {
		pubhash, err := common.ParsePublicHash(k)
		if err != nil {
			panic(err)
		}
		SeedNodeMap[pubhash] = netAddr
	}

	MaxBlocksPerFormulator := uint32(10)
	ChainID := uint8(0x01)
	Name := "FLEAT Mainnet"
	Version := uint16(0x0001)

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

	var back backend.StoreBackend
	var cdb *pile.DB
	switch cfg.BackendVersion {
	case 0:
		contextDB, err := backend.Create("badger", cfg.StoreRoot)
		if err != nil {
			panic(err)
		}
		back = contextDB
	case 1:
		contextDB, err := backend.Create("buntdb", cfg.StoreRoot+"/context")
		if err != nil {
			panic(err)
		}
		chainDB, err := pile.Open(cfg.StoreRoot + "/chain")
		if err != nil {
			panic(err)
		}
		chainDB.SetSyncMode(true)
		back = contextDB
		cdb = chainDB
	}
	st, err := chain.NewStore(back, cdb, ChainID, Name, Version)
	if err != nil {
		panic(err)
	}
	cm.Add("store", st)

	if st.Height() > 0 {
		if _, err := cdb.GetData(st.Height(), 0); err != nil {
			panic(err)
		}
	}

	cs := pof.NewConsensus(MaxBlocksPerFormulator, ObserverKeys)
	cs.SetMaxPhaseDiff(func(Height uint32) uint32 {
		if Height > 12270472 {
			return 2
		} else {
			return 0
		}
	})
	app := app.NewFletaApp()
	cn := chain.NewChain(cs, app, st)
	cn.MustAddProcess(admin.NewAdmin(1))
	cn.MustAddProcess(vault.NewVault(2))
	cn.MustAddProcess(formulator.NewFormulator(3))
	cn.MustAddProcess(gateway.NewGateway(4))
	cn.MustAddProcess(payment.NewPayment(5))
	as := apiserver.NewAPIServer()
	cn.MustAddService(as)
	if err := cn.Init(); err != nil {
		panic(err)
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

	nd := p2p.NewNode(ndkey, SeedNodeMap, cn, cfg.StoreRoot+"/peer")
	if err := nd.Init(); err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("node", nd)

	go nd.Run(":" + strconv.Itoa(cfg.Port))
	go as.Run(":" + strconv.Itoa(cfg.APIPort))

	cm.Wait()
}
