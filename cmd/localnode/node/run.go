package main

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
	"github.com/meverselabs/meverse/p2p"
)

func main() {
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
	ObserverKeys := make([]common.PublicKey, 0, len(obstrs))
	for _, v := range obstrs {
		if bs, err := hex.DecodeString(v); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(ChainID, bs); err != nil {
			panic(err)
		} else {
			obkeys = append(obkeys, Key)
			pubkey := Key.PublicKey()
			ObserverKeys = append(ObserverKeys, pubkey)
		}
	}

	var InitHash hash.Hash256
	var InitHeight uint32
	var InitTimestamp uint64

	// for i, obkey := range obkeys {
	// 	cdb, err := piledb.Open("../_test/odata_"+strconv.Itoa(i)+"/chain", InitHash, InitHeight, InitTimestamp)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	cdb.SetSyncMode(true)
	// 	st, err := chain.NewStore("../_test/odata_"+strconv.Itoa(i)+"/context", cdb, ChainID, Version)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	defer st.Close()

	// 	cn := chain.NewChain(ObserverKeys, st, "")
	// 	if err := cn.Init(localapp.Genesis()); err != nil {
	// 		panic(err)
	// 	}

	// 	ob := node.NewObserverNode(ChainID, obkey, NetAddressMap, cn, fmt.Sprintf("%v", i))
	// 	if err := ob.Init(); err != nil {
	// 		panic(err)
	// 	}

	// 	go ob.Run(":400"+strconv.Itoa(i), ":500"+strconv.Itoa(i))
	// }

	NdNetAddressMap := map[common.PublicKey]string{}

	frstrs := []string{
		"b000000000000000000000000000000000000000000000000000000000000010",
	}
	for i, v := range frstrs {
		if bs, err := hex.DecodeString(v); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(ChainID, bs); err != nil {
			panic(err)
		} else {
			pubkey := Key.PublicKey()
			NdNetAddressMap[pubkey] = ":600" + strconv.Itoa(i)
		}
	}

	ndstrs := []string{
		"b000000000000000000000000000000000000000000000000000000000000999",
		"b000000000000000000000000000000000000000000000000000000000000998",
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

	for i, ndkey := range ndkeys {
		if i == 0 {
			continue
		}
		cdb, err := piledb.Open("../_test/ndata_"+strconv.Itoa(i)+"/chain", InitHash, InitHeight, InitTimestamp)
		if err != nil {
			panic(err)
		}
		cdb.SetSyncMode(true)
		st, err := chain.NewStore("../_test/ndata_"+strconv.Itoa(i)+"/context", cdb, ChainID, Version)
		if err != nil {
			panic(err)
		}
		defer st.Close()

		cn := chain.NewChain(ObserverKeys, st, "")

		if err := cn.Init(localapp.Genesis()); err != nil {
			panic(err)
		}

		nd := p2p.NewNode(ChainID, ndkey, NdNetAddressMap, cn, "../_test/ndata_"+strconv.Itoa(i)+"/peer")
		if err := nd.Init(); err != nil {
			panic(err)
		}

		go nd.Run(":601" + strconv.Itoa(i))
	}

	select {}
}
