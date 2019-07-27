package main

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fletaio/fleta/encoding"

	"github.com/dgraph-io/badger"
	"github.com/fletaio/fleta/cmd/app"
	"github.com/fletaio/fleta/cmd/closer"
	"github.com/fletaio/fleta/cmd/config"
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/key"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/pof"
	"github.com/fletaio/fleta/process/admin"
	"github.com/fletaio/fleta/process/formulator"
	"github.com/fletaio/fleta/process/gateway"
	"github.com/fletaio/fleta/process/vault"
	"github.com/fletaio/fleta/service/apiserver"
	"github.com/fletaio/fleta/service/p2p"
)

// Config is a configuration for the cmd
type Config struct {
	SeedNodeMap  map[string]string
	NodeKeyHex   string
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

	st, err := chain.NewStore(cfg.StoreRoot, ChainID, Name, Version, cfg.ForceRecover)
	if err != nil {
		if cfg.ForceRecover || err != badger.ErrTruncateNeeded {
			panic(err)
		} else {
			fmt.Println(err)
			fmt.Println("Do you want to recover database(it can be failed)? [y/n]")
			var answer string
			fmt.Scanf("%s", &answer)
			if strings.ToLower(answer) == "y" {
				if s, err := chain.NewStore(cfg.StoreRoot, ChainID, Name, Version, true); err != nil {
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
	cn.MustAddProcess(admin.NewAdmin(1))
	cn.MustAddProcess(vault.NewVault(2))
	cn.MustAddProcess(formulator.NewFormulator(3))
	cn.MustAddProcess(gateway.NewGateway(4))
	as := apiserver.NewAPIServer()
	cn.MustAddService(as)
	cn.MustAddService(NewStatisticService())
	if err := cn.Init(); err != nil {
		panic(err)
	}
	cm.RemoveAll()
	cm.Add("chain", cn)

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

// StatisticService aggregates the chain and block data
type StatisticService struct {
	*types.ServiceBase
	pm                     types.ProcessManager
	cn                     types.Provider
	RangeBlockTimestampMap map[uint32]uint64
	RangeBlockCountMap     map[uint32]int
	RangeBlockSizeMap      map[uint32]int
}

// NewStatisticService returns a StatisticService
func NewStatisticService() *StatisticService {
	return &StatisticService{
		RangeBlockTimestampMap: map[uint32]uint64{},
		RangeBlockCountMap:     map[uint32]int{},
		RangeBlockSizeMap:      map[uint32]int{},
	}
}

// Name returns the name of the service
func (s *StatisticService) Name() string {
	return "fleta.statistics"
}

// Init initializes the service
func (s *StatisticService) Init(pm types.ProcessManager, cn types.Provider) error {
	s.pm = pm
	s.cn = cn
	return nil
}

// OnLoadChain called when the chain loaded
func (s *StatisticService) OnLoadChain(loader types.Loader) error {
	return nil
}

// OnBlockConnected called when a block is connected to the chain
func (s *StatisticService) OnBlockConnected(b *types.Block, events []types.Event, loader types.Loader) {
	idx := b.Header.Height / 86400
	s.RangeBlockCountMap[idx]++

	if b.Header.Height > 1 {
		bh, err := s.cn.Header(b.Header.Height - 1)
		if err != nil {
			panic(err)
		}
		s.RangeBlockTimestampMap[idx] += b.Header.Timestamp - bh.Timestamp
	} else {
		s.RangeBlockTimestampMap[idx] = 0
	}

	bs, err := encoding.Marshal(b)
	if err != nil {
		panic(err)
	}
	s.RangeBlockSizeMap[idx] += len(bs)

	for day, cnt := range s.RangeBlockCountMap {
		tsum := s.RangeBlockTimestampMap[day]
		ssum := s.RangeBlockSizeMap[day]
		log.Println(day, cnt, time.Duration(tsum/uint64(cnt)), ssum/cnt, 141338)
	}
}
