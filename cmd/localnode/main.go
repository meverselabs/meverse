package wallet

import (
	"encoding/hex"
	"math/big"
	"strconv"

	"github.com/meverselabs/meverse/cmd/localnode/localapp"
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/piledb"
	"github.com/meverselabs/meverse/node"
	"github.com/meverselabs/meverse/service/apiserver"
	"github.com/meverselabs/meverse/service/apiserver/metamaskrelay"
	"github.com/meverselabs/meverse/service/apiserver/viewchain"
	"github.com/meverselabs/meverse/service/bloomservice"
	"github.com/meverselabs/meverse/service/txsearch"
)

func Run() {
	ChainID := big.NewInt(0xFFFF)
	Version := uint16(0x0001)

	obstrs := []string{
		"c000000000000000000000000000000000000000000000000000000000000000",
		"c000000000000000000000000000000000000000000000000000000000000001",
		"c000000000000000000000000000000000000000000000000000000000000002",
		"c000000000000000000000000000000000000000000000000000000000000003",
		"c000000000000000000000000000000000000000000000000000000000000004",
	}
	obkeys := make([]key.Key, 0, len(obstrs))
	NetAddressMap := map[common.PublicKey]string{}
	FrNetAddressMap := map[common.PublicKey]string{}
	ObserverKeys := make([]common.PublicKey, 0, len(obstrs))
	for i, v := range obstrs {
		if bs, err := hex.DecodeString(v); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(ChainID, bs); err != nil {
			panic(err)
		} else {
			obkeys = append(obkeys, Key)
			pubkey := Key.PublicKey()
			ObserverKeys = append(ObserverKeys, pubkey)
			NetAddressMap[pubkey] = ":400" + strconv.Itoa(i)
			FrNetAddressMap[pubkey] = "ws://localhost:500" + strconv.Itoa(i)
		}
	}

	var InitHash hash.Hash256
	var InitHeight uint32
	var InitTimestamp uint64

	NdNetAddressMap := map[common.PublicKey]string{}

	frstrs := []string{
		"b000000000000000000000000000000000000000000000000000000000000010",
	}
	frkeys := make([]key.Key, 0, len(frstrs))
	for i, v := range frstrs {
		if bs, err := hex.DecodeString(v); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(ChainID, bs); err != nil {
			panic(err)
		} else {
			frkeys = append(frkeys, Key)
			pubkey := Key.PublicKey()
			NdNetAddressMap[pubkey] = ":600" + strconv.Itoa(i)
		}
	}

	ndstrs := []string{
		"b000000000000000000000000000000000000000000000000000000000000999",
	}
	ndkeys := make([]key.Key, 0, len(ndstrs))
	for i, v := range ndstrs {
		if bs, err := hex.DecodeString(v); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(ChainID, bs); err != nil {
			panic(err)
		} else {
			ndkeys = append(ndkeys, Key)
			pubkey := Key.PublicKey()
			NdNetAddressMap[pubkey] = ":601" + strconv.Itoa(i)
		}
	}

	for i, frkey := range frkeys {
		cdb, err := piledb.Open("../_test/fdata_"+strconv.Itoa(i)+"/chain", InitHash, InitHeight, InitTimestamp)
		if err != nil {
			panic(err)
		}
		cdb.SetSyncMode(true)
		st, err := chain.NewStore("../_test/fdata_"+strconv.Itoa(i)+"/context", cdb, ChainID, Version)
		if err != nil {
			panic(err)
		}
		defer st.Close()

		cn := chain.NewChain(ObserverKeys, st, "")

		var rpcapi *apiserver.APIServer
		var ts *txsearch.TxSearch
		if i == 0 {
			rpcapi = apiserver.NewAPIServer()
			ts = txsearch.NewTxSearch("../_test/_txsearch", rpcapi, st, cn, InitHeight)
			cn.MustAddService(ts)
			cn.MustAddService(rpcapi)
		}

		if err := cn.Init(localapp.Genesis()); err != nil {
			panic(err)
		}

		fr := node.NewGeneratorNode(ChainID, &node.GeneratorConfig{
			MaxTransactionsPerBlock: 10000,
		}, cn, frkey, frkey, FrNetAddressMap, NdNetAddressMap, "../_test/fdata_"+strconv.Itoa(i)+"/peer")
		if err := fr.Init(); err != nil {
			panic(err)
		}
		if i == 0 {
			metamaskrelay.NewMetamaskRelay(rpcapi, ts, &bloomservice.BloomBitService{}, cn, fr)
			go rpcapi.Run(":8541")
		}
		viewchain.NewViewchain(rpcapi, ts, cn, st, fr)

		go fr.Run(":600" + strconv.Itoa(i))
	}

	select {}
}
