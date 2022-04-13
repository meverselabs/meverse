package chain

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"testing"
	"time"

	"github.com/fletaio/fleta_v2/common"
	"github.com/fletaio/fleta_v2/common/amount"
	"github.com/fletaio/fleta_v2/common/bin"
	"github.com/fletaio/fleta_v2/common/hash"
	"github.com/fletaio/fleta_v2/common/key"
	"github.com/fletaio/fleta_v2/common/testapp"
	"github.com/fletaio/fleta_v2/core/piledb"
	"github.com/fletaio/fleta_v2/core/types"
)

func TestExecuteContractTx(t *testing.T) {
	signer := common.HexToAddress("0x477C578843cBe53C3568736347f640c2cdA4616F")
	token1Addr := common.HexToAddress("0xa1f093A1d8D4Ed5a7cC8fE29586266C5609a23e8")
	token2Addr := common.HexToAddress("0xE08FBAd440dfF3267f5A42061D64FC3b953C1Ef1")
	FormulatorAddr := common.HexToAddress("0xBaa3C856fbA6FFAda189D6bD0a89d5ef7959c75E")
	GeneratorAddr := common.HexToAddress("0x4dD2bf28E72EA48f83d9d3F398a03bF8baa8cC26")

	admin := common.HexToAddress("0x477C578843cBe53C3568736347f640c2cdA4616F")
	testAddr := common.HexToAddress("0xc42024AE9a4FAD398322d39E7e9aAb61bc5c6fe1")
	log.Println(signer, token1Addr, token2Addr, FormulatorAddr, admin, testAddr)

	tokenID1 := common.HexToAddress("0x6473217e48ca243ECDc38Dc927ad86B4BF7BCEC6")
	tokenID2 := common.HexToAddress("0xcF6f145d5C1dD40bD5943205CE2f31f8F8E3079f")
	tokenID3 := common.HexToAddress("0x37e1aC72838cb8F5432Dc0564039086D224b432E")
	tokenID4 := common.HexToAddress("0xa4F047659753871D458f77816B048236DEA10559")
	tokenID5 := common.HexToAddress("0x2D460Aa52e9A55d60f87cbe469a88AcD72Bd2EA7")
	tokenID6 := common.HexToAddress("0x66b828ABA72B4096a0fBB07fA59499198b48cC6e")
	tokenID7 := common.HexToAddress("0xE8aa351b01e304875A88D43163bfC120B796F086")
	tokenID8 := common.HexToAddress("0xcf7eDC8A6a3d72B48e05678b55Ae2d9a79773b2F")
	tokenID9 := common.HexToAddress("0xf0A016766c3F3E89440c5C29d28fF5A4D1983b7B")
	tokenID10 := common.HexToAddress("0x81B9fA17475E62180db827762D5e31376a1789A9")
	log.Println(tokenID1, tokenID2, tokenID3, tokenID4, tokenID5, tokenID6, tokenID7, tokenID8, tokenID9, tokenID10)

	newFormultors := []common.Address{}

	ctx, st := getTestContext()
	defer func() {
		st.Close()
		time.Sleep(time.Second)
		// err := os.RemoveAll("./_test")
		// if err != nil {
		// 	panic(err)
		// }
	}()

	type args struct {
		tx            *types.Transaction
		signer        common.Address
		printType     string
		isCreateAlpha bool
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		//mint test
		{name: "totalSupplyTx", args: args{tx: TotalSupplyTx(ctx, token1Addr), signer: signer, printType: "totalSupply"}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, token1Addr, admin), signer: signer, printType: "balanceOf"}},
		{name: "minttx", args: args{tx: MintTx(ctx, token1Addr, admin, amount.NewAmount(400000, 0)), signer: signer}},
		{name: "totalSupplyTx", args: args{tx: TotalSupplyTx(ctx, token1Addr), signer: signer, printType: "totalSupply"}},
		{name: "minttx", args: args{tx: MintTx(ctx, token1Addr, testAddr, amount.NewAmount(200000, 0)), signer: signer}},
		{name: "totalSupplyTx", args: args{tx: TotalSupplyTx(ctx, token1Addr), signer: signer, printType: "totalSupply"}},
		{name: "Burntx1", args: args{tx: BurnTx(ctx, token1Addr, amount.NewAmount(400000, 0)), signer: signer}},
		{name: "BurntxErr", args: args{tx: BurnTx(ctx, token1Addr, amount.NewAmount(200000, 0)), signer: testAddr}, wantErr: true},
		{name: "Burntx3", args: args{tx: BurnTx(ctx, token1Addr, amount.MustParseAmount("199999.999")), signer: testAddr}},
		{name: "Burntx3", args: args{tx: BurnTx(ctx, token1Addr, amount.MustParseAmount("0.001")), signer: signer}},
		{name: "totalSupplyTx", args: args{tx: TotalSupplyTx(ctx, token1Addr), signer: signer, printType: "totalSupply"}},

		//transfer test
		{name: "transfer", args: args{tx: TranferTx(ctx, token1Addr, testAddr), signer: signer}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, token1Addr, admin), signer: signer, printType: "balanceOf"}},
		//maintoken transfer test
		{name: "mainTransfer", args: args{tx: MainTokenTranferTx(ctx, testAddr), signer: signer}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, token1Addr, admin), signer: signer, printType: "balanceOf"}},

		//mint transferFrom test
		{name: "transferFrom", args: args{tx: TransferFromTx(ctx, token1Addr, admin, testAddr), signer: testAddr}, wantErr: true},
		{name: "approve", args: args{tx: ApproveTx(ctx, token1Addr, testAddr, amount.NewAmount(2000000, 0)), signer: signer}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, token1Addr, admin), signer: signer, printType: "balanceOf"}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, token1Addr, testAddr), signer: signer, printType: "balanceOf"}},
		{name: "transferFrom", args: args{tx: TransferFromTx(ctx, token1Addr, admin, testAddr), signer: testAddr}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, token1Addr, admin), signer: signer, printType: "balanceOf"}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, token1Addr, testAddr), signer: signer, printType: "balanceOf"}},

		//stake test
		{name: "StakingCheckTx", args: args{tx: StakingCheckTx(ctx, FormulatorAddr, GeneratorAddr, admin), signer: signer, printType: "StakingCheckTx"}},
		{name: "stake", args: args{tx: StakeTx(ctx, FormulatorAddr, GeneratorAddr, amount.NewAmount(300000, 0)), signer: signer}, wantErr: true},
		{name: "approve", args: args{tx: ApproveTx(ctx, token1Addr, FormulatorAddr, amount.NewAmount(200000, 0)), signer: signer}},
		{name: "allowance", args: args{tx: Allowance(ctx, token1Addr, admin, FormulatorAddr), signer: signer, printType: "Allowance"}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, token1Addr, admin), signer: signer, printType: "balanceOf"}},
		{name: "stake", args: args{tx: StakeTx(ctx, FormulatorAddr, GeneratorAddr, amount.NewAmount(200000, 0)), signer: signer}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, token1Addr, admin), signer: signer, printType: "balanceOf"}},
		{name: "StakingCheckTx", args: args{tx: StakingCheckTx(ctx, FormulatorAddr, GeneratorAddr, admin), signer: signer, printType: "StakingCheckTx"}},
		{name: "Unstake", args: args{tx: UnstakeTx(ctx, FormulatorAddr, GeneratorAddr, amount.NewAmount(300000, 0)), signer: signer}, wantErr: true},
		{name: "Unstake", args: args{tx: UnstakeTx(ctx, FormulatorAddr, GeneratorAddr, amount.NewAmount(100000, 0)), signer: signer}},
		{name: "StakingCheckTx", args: args{tx: StakingCheckTx(ctx, FormulatorAddr, GeneratorAddr, admin), signer: signer, printType: "StakingCheckTx"}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, token1Addr, admin), signer: signer, printType: "balanceOf"}},

		//create formulator test
		{name: "createAlpha", args: args{tx: CreateAlphaTx(ctx, FormulatorAddr), signer: signer, isCreateAlpha: true}, wantErr: true},
		{name: "approve", args: args{tx: ApproveTx(ctx, token1Addr, FormulatorAddr, amount.NewAmount(2000000, 0)), signer: signer}},
		{name: "allowen", args: args{tx: Allowance(ctx, token1Addr, admin, FormulatorAddr), signer: signer, printType: "Allowance"}},
		{name: "createAlpha", args: args{tx: CreateAlphaTx(ctx, FormulatorAddr), signer: signer, isCreateAlpha: true}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, token1Addr, admin), signer: signer, printType: "balanceOf"}},
		{name: "createAlphaBatch", args: args{tx: CreateAlphaBatchTx(ctx, FormulatorAddr, big.NewInt(9)), signer: signer, isCreateAlpha: true}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, FormulatorAddr, admin), signer: signer, printType: "balanceOf"}},

		//upgrade formulator test
		{name: "CreateSigma", args: args{tx: CreateSigmaTx(ctx, FormulatorAddr, []common.Address{
			tokenID1, tokenID2, tokenID3, tokenID1,
		}), signer: signer}, wantErr: true},
		{name: "CreateSigma", args: args{tx: CreateSigmaTx(ctx, FormulatorAddr, []common.Address{
			tokenID1, tokenID2, tokenID3, tokenID4,
		}), signer: signer}},

		//revoke formulator test
		{name: "totalSupplyTx", args: args{tx: TotalSupplyTx(ctx, token1Addr), signer: signer, printType: "totalSupply"}},
		{name: "revokeFormulate", args: args{tx: RevokeTx(ctx, FormulatorAddr, tokenID2), signer: signer}, wantErr: true},
		{name: "revokeFormulate", args: args{tx: RevokeTx(ctx, FormulatorAddr, tokenID5), signer: signer}},
		{name: "totalSupplyTx", args: args{tx: TotalSupplyTx(ctx, token1Addr), signer: signer, printType: "totalSupply"}},

		//sales test
		{name: "allowen", args: args{tx: Allowance(ctx, token1Addr, admin, FormulatorAddr), signer: signer, printType: "Allowance"}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, token1Addr, admin), signer: signer, printType: "balanceOf"}},
		{name: "OwnerOf", args: args{tx: OwnerOfTx(ctx, FormulatorAddr, tokenID1), signer: signer}},
		{name: "RegisterSales", args: args{tx: RegisterSalesTx(ctx, FormulatorAddr, tokenID1, amount.NewAmount(180000, 0)), signer: signer}},
		{name: "approve", args: args{tx: ApproveTx(ctx, token1Addr, FormulatorAddr, amount.NewAmount(170000, 0)), signer: testAddr}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, token1Addr, testAddr), signer: testAddr, printType: "balanceOf"}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, FormulatorAddr, testAddr), signer: testAddr, printType: "balanceOf"}},
		{name: "BuyFormulator1", args: args{tx: BuyFormulatorTx(ctx, FormulatorAddr, tokenID1), signer: testAddr}, wantErr: true},
		{name: "approve", args: args{tx: ApproveTx(ctx, token1Addr, FormulatorAddr, amount.NewAmount(180000, 0)), signer: testAddr}},
		{name: "BuyFormulator2", args: args{tx: BuyFormulatorTx(ctx, FormulatorAddr, tokenID1), signer: testAddr}, wantErr: false},
		{name: "OwnerOf", args: args{tx: OwnerOfTx(ctx, FormulatorAddr, tokenID1), signer: signer}},
		{name: "approve", args: args{tx: ApproveTx(ctx, token1Addr, FormulatorAddr, amount.NewAmount(180000, 0)), signer: testAddr}},
		{name: "BuyFormulator3", args: args{tx: BuyFormulatorTx(ctx, FormulatorAddr, tokenID1), signer: signer}, wantErr: true},
		{name: "OwnerOf", args: args{tx: OwnerOfTx(ctx, FormulatorAddr, tokenID1), signer: signer}},
		{name: "balanceOfTx", args: args{tx: BalanceOfTx(ctx, FormulatorAddr, testAddr), signer: testAddr, printType: "balanceOf"}},
		{name: "totalSupplyTx", args: args{tx: TotalSupplyTx(ctx, token1Addr), signer: signer, printType: "totalSupply"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fmt.Println("============ Start test", t.Name())

			sn := ctx.Snapshot()
			var err error
			var en *types.Event
			if en, err = ExecuteContractTx(ctx, tt.args.tx, tt.args.signer); err != nil {
				ctx.Revert(sn)
			} else {
				ctx.Commit(sn)
			}

			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteContractTx error = %+v, wantErr %v", err, tt.wantErr)
			} else if tt.wantErr {
				fmt.Printf("ExecuteContractTx error = %+v\n", err)
			} else {
				if en != nil {
					vs, err := bin.TypeReadAll(en.Result, -1)
					if err != nil {
						t.Errorf("ExecuteContractTx error = %+v", err)
					} else {
						if len(vs) > 0 {
							if am, ok := vs[0].(*amount.Amount); ok {
								is, err := bin.TypeReadAll(tt.args.tx.Args, 1)
								if err != nil {
									fmt.Println("err:", err)
								}
								if len(is) > 0 {
									addr, ok := is[0].(common.Address)
									if ok {
										fmt.Println(tt.args.printType, "addr:", addr, "amount:", am.String())
									} else {
										fmt.Println(tt.args.printType, "amount:", am.String())
									}
								} else {
									fmt.Println(tt.args.printType, "amount:", am.String())
								}
							} else {
								if tt.args.isCreateAlpha {
									if a, ok := vs[0].(common.Address); ok {
										newFormultors = append(newFormultors, a)
									} else if as, ok := vs[0].([]common.Address); ok {
										newFormultors = append(newFormultors, as...)
									}
									for _, a := range newFormultors {
										log.Println(a.String())
									}
								}
								fmt.Println("raw data:", vs[0])
							}
						} else {
							fmt.Println("return none")
						}
					}
				}

			}
			fmt.Println("============ end test", t.Name())
		})
	}

}

func CreateAlphaTx(ctx *types.Context, FormulatorAddr common.Address) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        FormulatorAddr,
		Method:    "CreateAlpha",
		Args:      nil,
	}
	return x
}

func CreateSigmaTx(ctx *types.Context, FormulatorAddr common.Address, tokenIDs []common.Address) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        FormulatorAddr,
		Method:    "CreateSigma",
		Args:      bin.TypeWriteAll(tokenIDs),
	}
	return x
}

func RevokeTx(ctx *types.Context, FormulatorAddr common.Address, tokenID common.Address) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        FormulatorAddr,
		Method:    "Revoke",
		Args:      bin.TypeWriteAll(tokenID),
	}
	return x
}

func CreateAlphaBatchTx(ctx *types.Context, FormulatorAddr common.Address, count *big.Int) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        FormulatorAddr,
		Method:    "CreateAlphaBatch",
		Args:      bin.TypeWriteAll(count),
	}
	return x
}

func ApproveTx(ctx *types.Context, token1Addr common.Address, FormulatorAddr common.Address, am *amount.Amount) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        token1Addr,
		Method:    "Approve",
		Args:      bin.TypeWriteAll(FormulatorAddr, am),
	}
	return x
}

func TranferTx(ctx *types.Context, token1Addr common.Address, testAddr common.Address) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        token1Addr,
		Method:    "Transfer",
		Args:      bin.TypeWriteAll(testAddr, amount.NewAmount(200000, 0)),
	}
	return x
}

func MainTokenTranferTx(ctx *types.Context, testAddr common.Address) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        testAddr,
		Method:    "Transfer",
		Args:      bin.TypeWriteAll(amount.NewAmount(200000, 0)),
	}
	return x
}

func TransferFromTx(ctx *types.Context, token1Addr common.Address, from common.Address, to common.Address) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        token1Addr,
		Method:    "TransferFrom",
		Args:      bin.TypeWriteAll(from, to, amount.NewAmount(2, 0)),
	}
	return x
}

func Allowance(ctx *types.Context, tokenAddr common.Address, owner common.Address, spender common.Address) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        tokenAddr,
		Method:    "Allowance",
		Args:      bin.TypeWriteAll(owner, spender),
	}
	return x
}

func RegisterSalesTx(ctx *types.Context, token1Addr common.Address, tokenID common.Address, am *amount.Amount) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        token1Addr,
		Method:    "RegisterSales",
		Args:      bin.TypeWriteAll(tokenID, am),
	}
	return x
}

func BuyFormulatorTx(ctx *types.Context, token1Addr common.Address, tokenID common.Address) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        token1Addr,
		Method:    "BuyFormulator",
		Args:      bin.TypeWriteAll(tokenID),
	}
	return x
}

func OwnerOfTx(ctx *types.Context, token1Addr common.Address, tokenID common.Address) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        token1Addr,
		Method:    "OwnerOf",
		Args:      bin.TypeWriteAll(tokenID),
	}
	return x
}

func TotalSupplyTx(ctx *types.Context, token1Addr common.Address) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        token1Addr,
		Method:    "TotalSupply",
		Args:      nil,
	}
	return x
}

func BalanceOfTx(ctx *types.Context, token1Addr common.Address, admin common.Address) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        token1Addr,
		Method:    "BalanceOf",
		Args:      bin.TypeWriteAll(admin),
	}
	return x
}

func BurnTx(ctx *types.Context, token1Addr common.Address, am *amount.Amount) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        token1Addr,
		Method:    "Burn",
		Args:      bin.TypeWriteAll(am),
	}
	return x
}

func MintTx(ctx *types.Context, token1Addr common.Address, admin common.Address, am *amount.Amount) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        token1Addr,
		Method:    "Mint",
		Args:      bin.TypeWriteAll(admin, am),
	}
	return x
}

func StakingCheckTx(ctx *types.Context, formulator, hyperAddr common.Address, stakingAddr common.Address) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        formulator,
		Method:    "StakingAmount",
		Args:      bin.TypeWriteAll(hyperAddr, stakingAddr),
	}
	return x
}

func StakeTx(ctx *types.Context, token1Addr common.Address, generator common.Address, am *amount.Amount) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        token1Addr,
		Method:    "Stake",
		Args:      bin.TypeWriteAll(generator, am),
	}
	return x
}

func UnstakeTx(ctx *types.Context, token1Addr common.Address, generator common.Address, am *amount.Amount) *types.Transaction {
	x := &types.Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: uint64(time.Now().UnixNano()),
		To:        token1Addr,
		Method:    "Unstake",
		Args:      bin.TypeWriteAll(generator, am),
	}
	return x
}

func getTestContext() (*types.Context, *Store) {
	var InitHash hash.Hash256
	var InitHeight uint32
	var InitTimestamp uint64
	ChainID := big.NewInt(0xd5)
	var Version uint16 = 1

	app := testapp.Genesis()
	cdb, err := piledb.Open("./_test/fdata_/chain", InitHash, InitHeight, InitTimestamp)
	if err != nil {
		panic(err)
	}
	cdb.SetSyncMode(true)
	st, err := NewStore("./_test/fdata_/context", cdb, ChainID, Version)
	if err != nil {
		panic(err)
	}

	obstrs := []string{
		"c000000000000000000000000000000000000000000000000000000000000000",
		"c000000000000000000000000000000000000000000000000000000000000001",
		"c000000000000000000000000000000000000000000000000000000000000002",
		"c000000000000000000000000000000000000000000000000000000000000003",
		"c000000000000000000000000000000000000000000000000000000000000004",
	}
	ObserverKeys := make([]common.PublicKey, 0, len(obstrs))
	for _, v := range obstrs {
		if bs, err := hex.DecodeString(v); err != nil {
			panic(err)
		} else if Key, err := key.NewMemoryKeyFromBytes(ChainID, bs); err != nil {
			panic(err)
		} else {
			ObserverKeys = append(ObserverKeys, Key.PublicKey())
		}
	}
	cn := NewChain(ObserverKeys, st, "")
	if err := cn.Init(app); err != nil {
		panic(err)
	}
	ctx := cn.NewContext()
	return ctx, st
}
