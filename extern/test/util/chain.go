package util

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/piledb"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"
)

func RemoveTestData() error {
	return os.RemoveAll("tdata/")
}

func RemoveChain(idx int) (string, error) {
	dir := "tdata/_data" + strconv.Itoa(idx)
	err := os.RemoveAll(dir)
	if err != nil {
		return "", err
	}
	return dir, nil
}

func (tc *TestContext) InitMainToken(adminAddress common.Address, ClassMap map[string]uint64) common.Address {
	arg := &token.TokenContractConstruction{
		Name:   "Test",
		Symbol: "TEST",
		InitialSupplyMap: map[common.Address]*amount.Amount{
			adminAddress: amount.NewAmount(2000000000, 0),
		},
	}
	bs, _, err := bin.WriterToBytes(arg)
	if err != nil {
		panic(err)
	}
	cont, err := tc.Ctx.DeployContract(adminAddress, ClassMap["Token"], bs)
	if err != nil {
		panic(err)
	}
	tokenAddress := cont.Address()
	tc.Ctx.SetMainToken(tokenAddress)
	// fmt.Println("Token", tokenAddress.String())
	return tokenAddress
}

func (tc *TestContext) InitChain(adm common.Address) error {
	dir, err := RemoveChain(tc.Idx)
	if err != nil {
		return err
	}

	Version := uint16(0x0001)

	cdb, err := piledb.Open(dir+"/chain", hash.Hash256{}, 0, 0)
	if err != nil {
		return err
	}

	cdb.SetSyncMode(true)
	st, err := chain.NewStore(dir+"/context", cdb, ChainID, Version)
	if err != nil {
		return err
	}
	cn := chain.NewChain(ObserverKeys, st, "main")
	if err := tc.Ctx.SetAdmin(adm, true); err != nil {
		return err
	}
	for _, v := range frkeys {
		if err := tc.Ctx.SetGenerator(v.PublicKey().Address(), true); err != nil {
			return err
		}
	}

	if err := cn.Init(tc.Ctx.Top()); err != nil {
		return err
	}

	if err := st.IterBlockAfterContext(func(b *types.Block) error {
		if err := cn.ConnectBlock(b, nil); err != nil {
			return err
		}
		return nil
	}); err != nil {
		if errors.Cause(err) == chain.ErrStoreClosed {
			return err
		}
		return err
	}

	tc.Cn = cn
	return nil
}

func (tc *TestContext) AddBlock(seconds uint64, txs []*types.Transaction, signer []key.Key) (hash.Hash256, error) {

	TimeoutCount := uint32(0)
	Generator, err := tc.Cn.TopGenerator(TimeoutCount)
	if err != nil {
		return hash.HexToHash(""), err
	}

	nextTimestamp := tc.Ctx.LastTimestamp() + (seconds * uint64(time.Second))
	bc := chain.NewBlockCreator(tc.Cn, tc.Ctx, Generator, TimeoutCount, nextTimestamp, 0)
	for i, tx := range txs {
		if tx != nil {
			sig, err := signer[i].Sign(tx.HashSig())
			if err != nil {
				return hash.HexToHash(""), err
			}

			err = bc.AddTx(tx, sig)
			if err != nil {
				return hash.HexToHash(""), err
			}
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

	err = tc.Cn.ConnectBlock(b, nil)
	if err != nil {
		return LastHash, err
	}

	return LastHash, nil
}

func (tc *TestContext) resetContext() {
	tc.Ctx = tc.Cn.NewContext()
}

func (tc *TestContext) SkipBlock(blockCount int) (*types.Context, error) {
	for i := 0; i < blockCount; i++ {
		_, err := tc.AddBlock(1, nil, nil)
		if err != nil {
			return nil, err
		}
		tc.resetContext()
	}
	return tc.Ctx, nil
}

func (tc *TestContext) MustSkipBlock(blockCount int) {
	for i := 0; i < blockCount; i++ {
		_, err := tc.AddBlock(1, nil, nil)
		if err != nil {
			panic(err)
		}
		tc.resetContext()
	}
}

func (tc *TestContext) Sleep(seconds uint64, tx []*types.Transaction, signer []key.Key) error {
	timestamp := tc.Ctx.LastTimestamp() + seconds*uint64(time.Second)

	LastHash, err := tc.AddBlock(seconds, tx, signer)
	if err != nil {
		tc.Ctx = tc.Cn.NewContext()
		return err
	}
	tc.Ctx = tc.Ctx.NextContext(LastHash, timestamp)
	return nil
}

func (tc *TestContext) SendTx(mkey key.Key, to common.Address, method string, params ...interface{}) ([]interface{}, error) {
	tx := &types.Transaction{
		ChainID:   ChainID,
		Timestamp: tc.Ctx.LastTimestamp(),
		To:        to,
		Method:    method,
	}

	tx.Args = bin.TypeWriteAll(params...)

	ins, err := bin.TypeReadAll(tx.Args, len(params))
	if err != nil {
		log.Println(ins, err)
		return nil, err
	}

	err = tc.Sleep(10, []*types.Transaction{tx}, []key.Key{mkey})
	if err != nil {
		return nil, err
	}
	b, err := tc.Cn.Provider().Block(tc.Ctx.TargetHeight() - 1)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(b.Body.Events); i++ {
		en := b.Body.Events[i]
		if en.Type == types.EventTagTxMsg {
			ins, err := bin.TypeReadAll(en.Result, 1)
			if err != nil {
				return nil, err
			}
			return ins, nil
		} else if en.Type == types.EventTagCallHistory {
			bf := bytes.NewBuffer(en.Result)
			mc := &types.MethodCallEvent{}
			mc.ReadFrom(bf)
		}
	}

	return nil, nil
}

func (tc *TestContext) ReadTx(mkey key.Key, to common.Address, method string, params ...interface{}) ([]interface{}, error) {
	tx := &types.Transaction{
		ChainID:   ChainID,
		Timestamp: tc.Ctx.LastTimestamp(),
		To:        to,
		Method:    method,
	}

	tx.Args = bin.TypeWriteAll(params...)

	ins, err := bin.TypeReadAll(tx.Args, len(params))
	if err != nil {
		log.Println(ins, err)
		return nil, err
	}

	n := tc.Ctx.Snapshot()
	ens, err := chain.ExecuteContractTxWithEvent(tc.Ctx, tx, mkey.PublicKey().Address(), "000000000000")
	tc.Ctx.Revert(n)
	if err != nil {
		return nil, err
	}

	for i := 0; i < len(ens); i++ {
		en := ens[i]
		if en.Type == types.EventTagTxMsg {
			ins, err := bin.TypeReadAll(en.Result, 1)
			if err != nil {
				return nil, err
			}
			return ins, nil
		} else if en.Type == types.EventTagCallHistory {
			bf := bytes.NewBuffer(en.Result)
			mc := &types.MethodCallEvent{}
			mc.ReadFrom(bf)
		}
	}
	return nil, nil
}

func (tc *TestContext) MakeTx(mkey key.Key, to common.Address, method string, params ...interface{}) ([]interface{}, error) {
	infs, err := tc.SendTx(mkey, to, method, params...)
	return infs, err
}

func (tc *TestContext) MustSendTx(mkey key.Key, to common.Address, method string, params ...interface{}) []interface{} {
	res, err := tc.SendTx(mkey, to, method, params...)
	if err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}
	return res
}

type TxCase struct {
	Mkey   key.Key
	To     common.Address
	Method string
	Params []interface{}
}

func (tc *TestContext) MustSendTxs(txcs []*TxCase) [][]interface{} {
	txs := []*types.Transaction{}
	ks := []key.Key{}
	for _, txc := range txcs {
		tx := &types.Transaction{
			ChainID:   ChainID,
			Timestamp: tc.Ctx.LastTimestamp(),
			To:        txc.To,
			Method:    txc.Method,
		}

		tx.Args = bin.TypeWriteAll(txc.Params...)

		ins, err := bin.TypeReadAll(tx.Args, len(txc.Params))
		if err != nil {
			log.Println(ins, err)
			panic(err)
		}
		txs = append(txs, tx)
		ks = append(ks, txc.Mkey)
	}

	err := tc.Sleep(10, txs, ks)
	if err != nil {
		panic(err)
	}
	b, err := tc.Cn.Provider().Block(tc.Ctx.TargetHeight() - 1)
	if err != nil {
		panic(err)
	}
	inss := [][]interface{}{}
	for i := 0; i < len(b.Body.Events); i++ {
		en := b.Body.Events[i]
		if en.Type == types.EventTagTxMsg {
			ins, err := bin.TypeReadAll(en.Result, 1)
			if err != nil {
				panic(err)
			}
			inss = append(inss, ins)
		}
	}

	return inss
}
