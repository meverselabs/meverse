package util

import (
	"bytes"
	"io"
	"reflect"
	"strings"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"
)

func Owner(cc *types.ContractContext, cont common.Address) (common.Address, error) {
	is, err := cc.Exec(cc, cont, "Owner", []interface{}{})
	if err != nil {
		return common.ZeroAddr, err
	}
	return is[0].(common.Address), nil
}

func (tc *TestContext) DeployContract(contType interface{}, contArgs io.WriterTo) common.Address {
	tx := &types.Transaction{
		ChainID:   ChainID,
		Timestamp: tc.Ctx.LastTimestamp() + 1,
		To:        common.ZeroAddr,
		Method:    "Contract.Deploy",
	}

	var classID uint64
	{
		rt := reflect.TypeOf(contType)
		for rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
		}
		name := rt.Name()
		if pkgPath := rt.PkgPath(); len(pkgPath) > 0 {
			pkgPath = strings.Replace(pkgPath, "meverselabs/meverse", "fletaio/fleta_v2", -1)
			name = pkgPath + "." + name
		}
		h := hash.Hash([]byte(name))
		classID = bin.Uint64(h[len(h)-8:])
	}

	dcd := chain.DeployContractData{
		Owner:   AdminKey.PublicKey().Address(),
		ClassID: classID,
	}

	{
		bf := &bytes.Buffer{}
		_, err := contArgs.WriteTo(bf)
		if err != nil {
			panic(err)
		}
		dcd.Args = bf.Bytes()
	}

	{
		bf := &bytes.Buffer{}
		_, err := dcd.WriteTo(bf)
		if err != nil {
			panic(err)
		}
		tx.Args = bf.Bytes()
	}
	err := tc.Sleep(1, []*types.Transaction{tx}, []key.Key{AdminKey})
	if err != nil {
		panic(err)
	}
	b, err := tc.Cn.Provider().Block(tc.Ctx.TargetHeight() - 1)
	if err != nil {
		panic(err)
	}
	en := b.Body.Events[0]
	ins, err := bin.TypeReadAll(en.Result, 1)
	if err != nil {
		panic(err)
	}
	tokenAddr := ins[0].(common.Address)
	return tokenAddr
}

func (tc *TestContext) MakeToken(name string, symbol string, amt string) common.Address {
	tokenContArgs := &token.TokenContractConstruction{
		Name:   name,
		Symbol: symbol,
		InitialSupplyMap: map[common.Address]*amount.Amount{
			AdminKey.PublicKey().Address(): amount.MustParseAmount(amt),
		},
	}
	tokenContType := &token.TokenContract{}

	tokenAddr := tc.DeployContract(tokenContType, tokenContArgs)
	return tokenAddr
}
