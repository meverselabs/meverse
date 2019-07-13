package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/dgraph-io/badger"
	"github.com/fletaio/fleta/cmd/app"
	"github.com/fletaio/fleta/cmd/closer"
	"github.com/fletaio/fleta/cmd/config"
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/pof"
	"github.com/fletaio/fleta/process/formulator"
	"github.com/fletaio/fleta/process/vault"
)

// Config is a configuration for the cmd
type Config struct {
	ObserverKeyMap map[string]string
	KeyHex         string
	ObseverPort    int
	FormulatorPort int
	APIPort        int
	StoreRoot      string
	ForceRecover   bool
}

func main() {
	var cfg Config
	if err := config.LoadFile("./config.toml", &cfg); err != nil {
		panic(err)
	}
	if len(cfg.StoreRoot) == 0 {
		cfg.StoreRoot = "./observer"
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

	MaxBlocksPerFormulator := uint32(10)
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

	st, err := chain.NewStore(cfg.StoreRoot, Name, Version, cfg.ForceRecover)
	if err != nil {
		if cfg.ForceRecover || err != badger.ErrTruncateNeeded {
			panic(err)
		} else {
			fmt.Println(err)
			fmt.Println("Do you want to recover database(it can be failed)? [y/n]")
			var answer string
			fmt.Scanf("%s", &answer)
			if strings.ToLower(answer) == "y" {
				if s, err := chain.NewStore(cfg.StoreRoot, Name, Version, true); err != nil {
					panic(err)
				} else {
					st = s
				}
			} else {
				os.Exit(1)
			}
		}
	}
	cm.Add("store", st)

	cs := pof.NewConsensus(MaxBlocksPerFormulator, ObserverKeys)
	app := app.NewFletaApp()
	cn := chain.NewChain(cs, app, st)
	cn.MustAddProcess(vault.NewVault(1))
	cn.MustAddProcess(formulator.NewFormulator(2, app.AdminAddress()))
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

	cm.Wait()
}
