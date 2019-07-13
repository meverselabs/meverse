package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/dgraph-io/badger"
	"github.com/fletaio/fleta/cmd/app"
	"github.com/fletaio/fleta/cmd/config"
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/pof"
	"github.com/fletaio/fleta/process/formulator"
	"github.com/fletaio/fleta/process/vault"
	"github.com/fletaio/fleta/service/p2p"
)

// Config is a configuration for the cmd
type Config struct {
	SeedNodeMap  map[string]string
	KeyHex       string
	ObserverKeys []string
	Port         int
	APIPort      int
	StoreRoot    string
	ForceRecover bool
}

func main() {
	var cfg Config
	if err := config.LoadFile("./config.toml", &cfg); err != nil {
		panic(err)
	}
	if len(cfg.StoreRoot) == 0 {
		cfg.StoreRoot = "./node"
	}

	var ndkey key.Key
	if len(cfg.KeyHex) > 0 {
		if bs, err := hex.DecodeString(cfg.KeyHex); err != nil {
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
	Name := "FLEAT Mainnet"
	Version := uint16(0x0001)

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
	cs := pof.NewConsensus(MaxBlocksPerFormulator, ObserverKeys)
	app := app.NewFletaApp()
	cn := chain.NewChain(cs, app, st)
	cn.MustAddProcess(vault.NewVault(1))
	cn.MustAddProcess(formulator.NewFormulator(2, app.AdminAddress()))
	if err := cn.Init(); err != nil {
		panic(err)
	}
	nd := p2p.NewNode(ndkey, SeedNodeMap, cn, cfg.StoreRoot+"/peer")
	if err := nd.Init(); err != nil {
		panic(err)
	}

	nd.Run(":" + strconv.Itoa(cfg.Port))
}
