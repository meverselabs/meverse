package testlib

import (
	"bytes"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/contract/exchange/factory"
	"github.com/meverselabs/meverse/contract/exchange/router"
	"github.com/meverselabs/meverse/contract/exchange/trade"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/contract/whitelist"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"
)

// DexInitialize deploy dex contracts and mint neccesary tokens to sigers
func DexInitialize(genesis *types.Context, classMap map[string]uint64, args []interface{}) ([]interface{}, error) {

	//alice(admin), bob, charlie
	alice, ok := args[0].(common.Address)
	if !ok {
		return nil, ErrArgument
	}
	bob, ok := args[1].(common.Address)
	if !ok {
		return nil, ErrArgument
	}
	charlie, ok := args[2].(common.Address)
	if !ok {
		return nil, ErrArgument
	}

	// factory
	factoryConstrunction := &factory.FactoryContractConstruction{Owner: alice}
	bs, _, err := bin.WriterToBytes(factoryConstrunction)
	if err != nil {
		return nil, err
	}
	v, err := genesis.DeployContract(alice, classMap["Factory"], bs)
	if err != nil {
		return nil, err
	}
	factory := v.(*factory.FactoryContract).Address()

	// router
	routerConstrunction := &router.RouterContractConstruction{Factory: factory}
	bs, _, err = bin.WriterToBytes(routerConstrunction)
	if err != nil {
		return nil, err
	}
	v, err = genesis.DeployContract(alice, classMap["Router"], bs)
	if err != nil {
		return nil, err
	}
	router := v.(*router.RouterContract).Address()

	// whitelist
	whitelistConstrunction := &whitelist.WhiteListContractConstruction{}
	bs, _, err = bin.WriterToBytes(whitelistConstrunction)
	if err != nil {
		return nil, err
	}
	v, err = genesis.DeployContract(alice, classMap["WhiteList"], bs)
	if err != nil {
		return nil, err
	}
	whiteList := v.(*whitelist.WhiteListContract).Address()

	// tokens
	size := uint8(2)
	uniTokens := make([]common.Address, size, size)
	for k := uint8(0); k < size; k++ {
		tokenConstrunction := &token.TokenContractConstruction{
			Name:   "Token" + strconv.Itoa(int(k)),
			Symbol: "TOKEN" + strconv.Itoa(int(k)),
			InitialSupplyMap: map[common.Address]*amount.Amount{
				alice:   amount.NewAmount(100000000, 0),
				bob:     amount.NewAmount(100000000, 0),
				charlie: amount.NewAmount(100000000, 0),
			},
		}

		bs, _, _ := bin.WriterToBytes(tokenConstrunction)
		v, _ := genesis.DeployContract(alice, classMap["Token"], bs)
		uniTokens[k] = v.(*token.TokenContract).Address()
	}

	// pair
	_PairName := "__UNI_NAME"
	_PairSymbol := "__UNI_SYMBOL"
	_Fee := uint64(40000000)
	_AdminFee := uint64(trade.MAX_ADMIN_FEE)
	_WinnerFee := uint64(5000000000)
	_GroupId := hash.BigToHash(big.NewInt(1))

	is, err := Exec(genesis, alice, factory, "CreatePairUni", []interface{}{uniTokens[0], uniTokens[1], common.Address{}, _PairName, _PairSymbol, alice, charlie, _Fee, _AdminFee, _WinnerFee, whiteList, _GroupId, classMap["UniSwap"]})
	if err != nil {
		return nil, err
	}
	pair := is[0].(common.Address)

	token0, token1, err := trade.SortTokens(uniTokens[0], uniTokens[1])
	if err != nil {
		return nil, err
	}

	// approve
	for _, token := range []common.Address{token0, token1} {
		for _, signer := range []common.Address{alice, bob, charlie} {
			TokenApprove(genesis, token, signer, router)
		}
	}

	fmt.Println("factory   Address", factory.String())
	fmt.Println("router    Address", router.String())
	fmt.Println("whitelist Address", whiteList.String())
	fmt.Println("token0    Address", token0.String())
	fmt.Println("token1    Address", token1.String())
	fmt.Println("pair      Address", pair.String())

	return []interface{}{&factory, &router, &whiteList, &token0, &token1, &pair}, nil
}

// // Go Contract Interface
// type IGoContract interface {
// 	//	GetAddress() *common.Address
// 	SetAddress(*common.Address)
// }

// NewEvmContract makes new Evm Contract
type GoContract struct {
	Address *common.Address
}

// SetAddress sets the address of GoContract
func (c *GoContract) SetAddress(addr *common.Address) {
	c.Address = addr
}

// MakeGoContract makes go Contract transaction
func MakeGoTx(senderKey key.Key, provider types.Provider, cont *common.Address, method string, args ...any) *TxWithSigner {
	return &TxWithSigner{
		Tx: &types.Transaction{
			ChainID:   provider.ChainID(),
			Timestamp: uint64(time.Now().UnixNano()),
			To:        *cont,
			Method:    method,
			Args:      bin.TypeWriteAll(args...),
		},
		Signer: senderKey,
	}
}

// Token is a Go Token Contract
type Token struct {
	GoContract
}

// NewGoContract makes new Go Contract
func NewTokenContract(address *common.Address) *Token {
	token := &Token{}
	token.SetAddress(address)
	return token
}

// NewTokenTx makes the token deploy tx
func NewTokenTx(tb *TestBlockChain, senderKey key.Key, name, symbol string, initialSupplyMap map[common.Address]*amount.Amount) (*TxWithSigner, error) {

	arg := &token.TokenContractConstruction{
		Name:             name,
		Symbol:           symbol,
		InitialSupplyMap: initialSupplyMap,
	}
	bs, _, _ := bin.WriterToBytes(arg)

	data := &chain.DeployContractData{
		Owner:   senderKey.PublicKey().Address(),
		ClassID: tb.ClassMap["Token"],
		Args:    bs,
	}

	bs2 := bytes.NewBuffer([]byte{})
	_, err := data.WriteTo(bs2)
	if err != nil {
		return nil, err
	}

	return &TxWithSigner{
		Tx: &types.Transaction{
			ChainID:   tb.ChainID,
			Timestamp: uint64(time.Now().UnixNano()),
			To:        common.Address{},
			Method:    "Contract.Deploy",
			Args:      bs2.Bytes(),
		},
		Signer: senderKey,
	}, nil
}

// TransferTx returns the Transfer method tx of the Token contract
func (c *Token) TransferTx(senderKey key.Key, provider types.Provider, to common.Address, amt *amount.Amount) *TxWithSigner {
	return MakeGoTx(senderKey, provider, c.Address, "Transfer", to, amt)
}

// ApproveTx returns the Approve method tx of the Token contract
func (c *Token) ApproveTx(senderKey key.Key, provider types.Provider, to common.Address, amt *amount.Amount) *TxWithSigner {
	return MakeGoTx(senderKey, provider, c.Address, "Approve", to, amt)
}

// Router is a Go Router Contract
type Router struct {
	GoContract
}

// NewGoContract makes new Go Contract
func NewRouterContract(address *common.Address) *Router {
	router := &Router{}
	router.SetAddress(address)
	return router
}

// UniAddLiquidityTx returns the UniAddLiquidity method tx of the Router contract
func (c *Router) UniAddLiquidityTx(senderKey key.Key, provider types.Provider, token0, token1 common.Address, amountADesired, amountBDesired, amountAMin, amountBMin *amount.Amount) *TxWithSigner {
	return MakeGoTx(senderKey, provider, c.Address, "UniAddLiquidity", token0, token1, amountADesired, amountBDesired, amountAMin, amountBMin)
}
