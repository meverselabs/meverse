package util

import (
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/piledb"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"
)

var (
	chainID  = big.NewInt(1)
	frKeyMap = map[common.Address]key.Key{}
	obkeys   = []key.Key{}
)

func Chain(idx int, ctx *types.Context, adm common.Address) (*chain.Chain, error) {
	dir, err := RemoveChain(idx)
	if err != nil {
		return nil, err
	}

	Version := uint16(0x0001)
	//obkeys := []key.Key{}
	ObserverKeys := []common.PublicKey{}
	for i := 0; i < 5; i++ {
		pk, err := key.NewMemoryKeyFromBytes(chainID, []byte{1, 1, byte(i), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
		if err != nil {
			return nil, err
		}
		obkeys = append(obkeys, pk)
		ObserverKeys = append(ObserverKeys, pk.PublicKey())
	}
	frkeys := []key.Key{}
	//frKeyMap := map[common.Address]key.Key{}
	for i := 0; i < 10; i++ {
		pk, err := key.NewMemoryKeyFromBytes(chainID, []byte{1, 1, 1, byte(i), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
		if err != nil {
			return nil, err
		}
		frkeys = append(frkeys, pk)
		frKeyMap[pk.PublicKey().Address()] = pk
	}

	cdb, err := piledb.Open(dir+"/chain", hash.Hash256{}, 0, 0)
	if err != nil {
		return nil, err
	}

	cdb.SetSyncMode(true)
	st, err := chain.NewStore(dir+"/context", cdb, chainID, Version)
	if err != nil {
		return nil, err
	}

	cn := chain.NewChain(ObserverKeys, st, "main")
	if err := ctx.SetAdmin(adm, true); err != nil {
		return nil, err
	}
	for _, v := range frkeys {
		if err := ctx.SetGenerator(v.PublicKey().Address(), true); err != nil {
			return nil, err
		}
	}

	if err := cn.Init(ctx.Top()); err != nil {
		return nil, err
	}

	if err := st.IterBlockAfterContext(func(b *types.Block) error {
		if err := cn.ConnectBlock(b, nil); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if errors.Cause(err) == chain.ErrStoreClosed {
			return nil, err
		}
		return nil, err
	}

	return cn, nil
}

// AfterSuite 에서 일괄 삭제
func RemoveChain(idx int) (string, error) {
	dir := "chain/_data" + strconv.Itoa(idx)
	if _, err := os.Stat("/mnt/ramdisk"); !os.IsNotExist(err) {
		dir = "/mnt/ramdisk/" + dir
	}
	/*
		err := os.RemoveAll(dir)
		if err != nil {
			return "", err
		}
	*/
	return dir, nil
}

func AddBlock(cn *chain.Chain, ctx *types.Context, tx *types.Transaction, signer key.Key) (hash.Hash256, error) {
	TimeoutCount := uint32(0)
	Generator, err := cn.TopGenerator(TimeoutCount)
	if err != nil {
		return hash.HexToHash(""), err
	}

	bc := chain.NewBlockCreator(cn, ctx, Generator, TimeoutCount, ctx.LastTimestamp(), 0)
	if tx != nil {
		sig, err := signer.Sign(tx.HashSig())
		if err != nil {
			return hash.HexToHash(""), err
		}
		if err := bc.AddTx(tx, sig); err != nil {
			return hash.HexToHash(""), err
		}
	}

	b, err := bc.Finalize(0)
	if err != nil {
		return hash.HexToHash(""), err
	}

	HeaderHash := bin.MustWriterToHash(&b.Header)

	LastHash := HeaderHash

	pk := frKeyMap[Generator]
	GenSig, err := pk.Sign(HeaderHash)
	if err != nil {
		return hash.HexToHash(""), err
	}

	b.Body.BlockSignatures = append(b.Body.BlockSignatures, GenSig)

	blockSign := &types.BlockSign{
		HeaderHash:         HeaderHash,
		GeneratorSignature: GenSig,
	}

	BlockSignHash := bin.MustWriterToHash(blockSign)

	idxes := rand.Perm(len(obkeys))
	for i := 0; i < len(obkeys)/2+1; i++ {
		pk := obkeys[idxes[i]]
		ObSig, err := pk.Sign(BlockSignHash)
		if err != nil {
			return hash.HexToHash(""), err
		}
		b.Body.BlockSignatures = append(b.Body.BlockSignatures, ObSig)
	}

	cn.ConnectBlock(b, nil)

	return LastHash, nil
}
func Sleep(cn *chain.Chain, ctx *types.Context, tx *types.Transaction, seconds uint64, signer key.Key) (*types.Context, error) {
	LastHash, err := AddBlock(cn, ctx, tx, signer)
	if err != nil {
		return nil, err
	}
	timestamp := ctx.LastTimestamp() + seconds*uint64(time.Second)
	return ctx.NextContext(LastHash, timestamp), nil
}
