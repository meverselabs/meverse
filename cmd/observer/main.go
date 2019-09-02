package main

import (
	"encoding/hex"
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
	"github.com/fletaio/fleta/pof"
	"github.com/fletaio/fleta/process/admin"
	"github.com/fletaio/fleta/process/formulator"
	"github.com/fletaio/fleta/process/gateway"
	"github.com/fletaio/fleta/process/payment"
	"github.com/fletaio/fleta/process/vault"
	"github.com/fletaio/fleta/service/apiserver"
)

// Config is a configuration for the cmd
type Config struct {
	ObserverKeyMap map[string]string
	KeyHex         string
	ObseverPort    int
	FormulatorPort int
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
		cfg.StoreRoot = "./odata"
	}
	if len(cfg.RLogHost) > 0 && cfg.UseRLog {
		if len(cfg.RLogPath) == 0 {
			cfg.RLogPath = "./odata_rlog"
		}
		rlog.SetRLogHost(cfg.RLogHost)
		rlog.Enablelogger(cfg.RLogPath)
	}

	var obkey key.Key
	if bs, err := hex.DecodeString(cfg.KeyHex); err != nil {
		panic(err)
	} else if Key, err := key.NewMemoryKeyFromBytes(bs); err != nil {
		panic(err)
	} else {
		obkey = Key
	}

	NetAddressMap := map[common.PublicHash]string{}
	ObserverKeys := []common.PublicHash{}
	for k, netAddr := range cfg.ObserverKeyMap {
		pubhash, err := common.ParsePublicHash(k)
		if err != nil {
			panic(err)
		}
		NetAddressMap[pubhash] = netAddr
		ObserverKeys = append(ObserverKeys, pubhash)
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

	MaxBlocksPerFormulator := uint32(10)
	ChainID := uint8(0x01)
	Name := "FLEAT Mainnet"
	Version := uint16(0x0001)

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

	cs := pof.NewConsensus(MaxBlocksPerFormulator, ObserverKeys)
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

	ob := pof.NewObserverNode(obkey, NetAddressMap, cs)
	if err := ob.Init(); err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("observer", ob)

	go ob.Run(":"+strconv.Itoa(cfg.ObseverPort), ":"+strconv.Itoa(cfg.FormulatorPort))
	go as.Run(":" + strconv.Itoa(cfg.APIPort))

	cm.Wait()
}
