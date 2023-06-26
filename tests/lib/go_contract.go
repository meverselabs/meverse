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
	"github.com/meverselabs/meverse/contract/formulator"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/contract/whitelist"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"
)

// DexInitialize deploy dex contracts and mint neccesary tokens to sigers
func DexInitialize(genesis *types.Context, classMap map[string]uint64, owner, winner common.Address, intialSupplyMap map[common.Address]*amount.Amount) (*common.Address, *common.Address, *common.Address, *common.Address, *common.Address, *common.Address, error) {

	// factory
	factoryConstrunction := &factory.FactoryContractConstruction{Owner: owner}
	bs, _, err := bin.WriterToBytes(factoryConstrunction)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	v, err := genesis.DeployContract(owner, classMap["Factory"], bs)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	factory := v.(*factory.FactoryContract).Address()

	// router
	routerConstrunction := &router.RouterContractConstruction{Factory: factory}
	bs, _, err = bin.WriterToBytes(routerConstrunction)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	v, err = genesis.DeployContract(owner, classMap["Router"], bs)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	router := v.(*router.RouterContract).Address()

	// whitelist
	whitelistConstrunction := &whitelist.WhiteListContractConstruction{}
	bs, _, err = bin.WriterToBytes(whitelistConstrunction)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	v, err = genesis.DeployContract(owner, classMap["WhiteList"], bs)
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	whiteList := v.(*whitelist.WhiteListContract).Address()

	// tokens
	size := uint8(2)
	uniTokens := make([]common.Address, size, size)
	for k := uint8(0); k < size; k++ {
		tokenConstrunction := &token.TokenContractConstruction{
			Name:             "Token" + strconv.Itoa(int(k)),
			Symbol:           "TOKEN" + strconv.Itoa(int(k)),
			InitialSupplyMap: intialSupplyMap,
		}

		bs, _, _ := bin.WriterToBytes(tokenConstrunction)
		v, _ := genesis.DeployContract(owner, classMap["Token"], bs)
		uniTokens[k] = v.(*token.TokenContract).Address()
	}

	// pair
	_PairName := "__UNI_NAME"
	_PairSymbol := "__UNI_SYMBOL"
	_Fee := uint64(40000000)
	_AdminFee := uint64(trade.MAX_ADMIN_FEE)
	_WinnerFee := uint64(5000000000)
	_GroupId := hash.BigToHash(big.NewInt(1))

	is, err := Exec(genesis, owner, factory, "CreatePairUni", []interface{}{uniTokens[0], uniTokens[1], common.Address{}, _PairName, _PairSymbol, owner, winner, _Fee, _AdminFee, _WinnerFee, whiteList, _GroupId, classMap["UniSwap"]})
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}
	pair := is[0].(common.Address)

	token0, token1, err := trade.SortTokens(uniTokens[0], uniTokens[1])
	if err != nil {
		return nil, nil, nil, nil, nil, nil, err
	}

	// approve
	supplyAddresses := []common.Address{}
	for k := range intialSupplyMap {
		supplyAddresses = append(supplyAddresses, k)
	}
	for _, token := range []common.Address{token0, token1} {
		for _, signer := range supplyAddresses {
			TokenApprove(genesis, token, signer, router)
		}
	}

	fmt.Println("factory   Address", factory.String())
	fmt.Println("router    Address", router.String())
	fmt.Println("whitelist Address", whiteList.String())
	fmt.Println("token0    Address", token0.String())
	fmt.Println("token1    Address", token1.String())
	fmt.Println("pair      Address", pair.String())

	return &factory, &router, &whiteList, &token0, &token1, &pair, nil
}

// // Go Contract Interface
// type IGoContract interface {
// 	//	GetAddress() *common.Address
// 	SetAddress(*common.Address)
// }

// GoContract
type GoContract struct {
	Address  *common.Address
	Provider types.Provider
}

// SetAddress sets the address of GoContract
func (c *GoContract) SetAddress(addr *common.Address) {
	c.Address = addr
}

// SetProvider sets the provider of GoContract
func (c *GoContract) SetProvider(provider types.Provider) {
	c.Provider = provider
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
type TokenContract struct {
	GoContract
}

// NewGoContract makes new Go Contract
func BindTokenContract(address *common.Address, provider types.Provider) *TokenContract {
	return &TokenContract{GoContract: GoContract{Address: address, Provider: provider}}
}

// DeployTokenTx makes the token deploy tx
func DeployTokenTx(tb *TestBlockChain, senderKey key.Key, name, symbol string, initialSupplyMap map[common.Address]*amount.Amount) (*TxWithSigner, error) {

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
func (c *TokenContract) TransferTx(senderKey key.Key, to common.Address, amt *amount.Amount) *TxWithSigner {
	return MakeGoTx(senderKey, c.Provider, c.Address, "Transfer", to, amt)
}

// ApproveTx returns the Approve method tx of the Token contract
func (c *TokenContract) ApproveTx(senderKey key.Key, to common.Address, amt *amount.Amount) *TxWithSigner {
	return MakeGoTx(senderKey, c.Provider, c.Address, "Approve", to, amt)
}

// Router is a Go Router Contract
type RouterContract struct {
	GoContract
}

// NewGoContract makes new Go Contract
func BindRouterContract(address *common.Address, provider types.Provider) *RouterContract {
	return &RouterContract{GoContract: GoContract{Address: address, Provider: provider}}
}

// UniAddLiquidityTx returns the UniAddLiquidity method tx of the Router contract
func (c *RouterContract) UniAddLiquidityTx(senderKey key.Key, token0, token1 common.Address, amountADesired, amountBDesired, amountAMin, amountBMin *amount.Amount) *TxWithSigner {
	return MakeGoTx(senderKey, c.Provider, c.Address, "UniAddLiquidity", token0, token1, amountADesired, amountBDesired, amountAMin, amountBMin)
}

// Formulator is a Go Formulator Contract
type FormulatorContract struct {
	GoContract
}

var DefaultFormulatorPolicy = &formulator.FormulatorPolicy{
	AlphaAmount:    amount.NewAmount(200000, 0),
	SigmaCount:     4,
	SigmaBlocks:    200,
	OmegaCount:     2,
	OmegaBlocks:    300,
	HyperAmount:    amount.NewAmount(3000000, 0),
	MinStakeAmount: amount.NewAmount(100, 0),
}

func formulatorContractConstruction(mev, admin common.Address) *formulator.FormulatorContractConstruction {
	return &formulator.FormulatorContractConstruction{
		TokenAddress:     mev,
		FormulatorPolicy: *DefaultFormulatorPolicy,
		RewardPolicy: formulator.RewardPolicy{
			RewardPerBlock:        amount.MustParseAmount("0.6341958396752917"),
			AlphaEfficiency1000:   1000,
			SigmaEfficiency1000:   1150,
			OmegaEfficiency1000:   1300,
			HyperEfficiency1000:   1300,
			StakingEfficiency1000: 700,
			CommissionRatio1000:   50,
			MiningFeeAddress:      admin,
			MiningFee1000:         300,
		},
	}
}

// DeployFormulatorTx makes the formulator deploy tx
func DeployFormulatorTx(tb *TestBlockChain, senderKey key.Key, mev, admin common.Address) (*TxWithSigner, error) {

	arg := formulatorContractConstruction(mev, admin)
	bs, _, err := bin.WriterToBytes(arg)
	if err != nil {
		return nil, err
	}
	data := &chain.DeployContractData{
		Owner:   senderKey.PublicKey().Address(),
		ClassID: tb.ClassMap["Formulator"],
		Args:    bs,
	}

	bs2 := bytes.NewBuffer([]byte{})
	_, err = data.WriteTo(bs2)
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

// NewFormulatorContract gets the depoloyed formulator contract by address
func BindFormulatorContract(address *common.Address, provider types.Provider) *FormulatorContract {
	return &FormulatorContract{GoContract: GoContract{Address: address, Provider: provider}}
}

// FormulatorInitialize deploy formulator contract and set alpha, sigma and omega Owners
func FormulatorInitialize(genesis *types.Context, classMap map[string]uint64, mev, admin common.Address, alphaOwners, sigmaOwners, omegaOwners []common.Address) (*common.Address, error) {

	arg := formulatorContractConstruction(mev, admin)
	bs, _, err := bin.WriterToBytes(arg)
	if err != nil {
		return nil, err
	}

	v, err := genesis.DeployContract(admin, classMap["Formulator"], bs)
	if err != nil {
		panic(err)
	}
	cont := v.(*formulator.FormulatorContract)
	formulatorAddress := cont.Address()

	if true {
		v, err := genesis.Contract(mev)
		if err != nil {
			panic(err)
		}
		cont := v.(*token.TokenContract)
		cc := genesis.ContractContext(cont, admin)
		if err := cont.SetMinter(cc, formulatorAddress, true); err != nil {
			panic(err)
		}
	}

	cc := genesis.ContractContext(cont, cont.Address())
	intr := types.NewInteractor(genesis, cont, cc, "000000000000", false)
	cc.Exec = intr.Exec
	for _, addr := range alphaOwners {
		if _, err := cont.CreateGenesisAlpha(cc, addr); err != nil {
			panic(err)
		}
	}
	for _, addr := range sigmaOwners {
		if _, err := cont.CreateGenesisSigma(cc, addr); err != nil {
			panic(err)
		}
	}
	for _, addr := range omegaOwners {
		if _, err := cont.CreateGenesisOmega(cc, addr); err != nil {
			panic(err)
		}
	}

	fmt.Println("formulator Address", formulatorAddress)

	return &formulatorAddress, nil
}

// CreateAlphaTx returns the  CreateAlpha method tx of the Formulator contract
func (c *FormulatorContract) CreateAlphaTx(senderKey key.Key) *TxWithSigner {
	return MakeGoTx(senderKey, c.Provider, c.Address, "CreateAlpha")
}

// CreateSigmaTx returns the  CreateSigma method tx of the Formulator contract
func (c *FormulatorContract) CreateSigmaTx(senderKey key.Key, tokenIDs []common.Address) *TxWithSigner {
	return MakeGoTx(senderKey, c.Provider, c.Address, "CreateSigma", tokenIDs)
}

// CreateOmegaTx returns the  CreateOmega method tx of the Formulator contract
func (c *FormulatorContract) CreateOmegaTx(senderKey key.Key, tokenIDs []common.Address) *TxWithSigner {
	return MakeGoTx(senderKey, c.Provider, c.Address, "CreateOmega", tokenIDs)
}

// SyncGeneratorTx returns the  SyncGenerator method tx of the Formulator contract
func (c *FormulatorContract) SyncGeneratorTx(senderKey key.Key) *TxWithSigner {
	return MakeGoTx(senderKey, c.Provider, c.Address, "SyncGenerator")
}

// RevokeTx returns the  Revoke method tx of the Formulator contract
func (c *FormulatorContract) RevokeTx(senderKey key.Key, tokenID common.Address) *TxWithSigner {
	return MakeGoTx(senderKey, c.Provider, c.Address, "Revoke", tokenID)
}

// TransferFromTx returns the  TransferFrom method tx of the Formulator contract
func (c *FormulatorContract) TransferFromTx(senderKey key.Key, from, to, tokenID common.Address) *TxWithSigner {
	return MakeGoTx(senderKey, c.Provider, c.Address, "TransferFrom", from, to, tokenID)
}

// RegisterSalesTx returns the  RegisterSales method tx of the Formulator contract
func (c *FormulatorContract) RegisterSalesTx(senderKey key.Key, tokenID common.Address, amt *amount.Amount) *TxWithSigner {
	return MakeGoTx(senderKey, c.Provider, c.Address, "RegisterSales", tokenID, amt)
}

// BuyFormulatorTx returns the  BuyFormulator method tx of the Formulator contract
func (c *FormulatorContract) BuyFormulatorTx(senderKey key.Key, tokenID common.Address) *TxWithSigner {
	return MakeGoTx(senderKey, c.Provider, c.Address, "BuyFormulator", tokenID)
}
