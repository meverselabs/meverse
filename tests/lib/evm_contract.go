package testlib

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/core/ctypes"
	"github.com/meverselabs/meverse/core/types"
	mtypes "github.com/meverselabs/meverse/ethereum/core/types"
	"github.com/meverselabs/meverse/service/pack"
)

// Evm Contract Interface
type IEvmContract interface {
	Set(abi *abi.ABI, provider types.Provider, addr *common.Address)
	SetAddress(*common.Address)
	makeEvmTx(senderKey key.Key, nonceIncrement uint64, data []byte) (*TxWithSigner, error)
}

// Evm Contract
type EvmContract struct {
	Abi      *abi.ABI
	Provider types.Provider
	Address  *common.Address
}

// SetAddress sets the address of EvmContract
func (c *EvmContract) Set(abi *abi.ABI, provider types.Provider, addr *common.Address) {
	c.Abi = abi
	c.Provider = provider
	c.Address = addr
}

// SetAddress sets the address of EvmContract
func (c *EvmContract) SetAddress(addr *common.Address) {
	c.Address = addr
}

// MakeEvmTx makes Evm Tx by argument([]any)
// @arg nonceIncrement : 한 signer가 동시여 여러개의 transaction을 실행할경우 nonceIncrement로 조정해 주어야 한다.
func (c *EvmContract) MakeEvmTx(senderKey key.Key, nonceIncrement uint64, method string, args ...any) (*TxWithSigner, error) {

	data, err := c.Abi.Pack(method, args...)
	if err != nil {
		return nil, err
	}

	return c.makeEvmTx(senderKey, nonceIncrement, data)

}

// MakeEvmTx makes Evm Tx by argument([]byte)
func (c *EvmContract) makeEvmTx(senderKey key.Key, nonceIncrement uint64, data []byte) (*TxWithSigner, error) {

	provider := c.Provider
	chainID := provider.ChainID()
	signer := mtypes.MakeSigner(chainID, provider.Height())

	txSigner := senderKey.PublicKey().Address()
	nonce := provider.AddrSeq(txSigner) + nonceIncrement
	etx := etypes.NewTx(&etypes.AccessListTx{
		ChainID:  chainID,
		Nonce:    nonce,
		Data:     data,
		To:       c.Address,
		Gas:      GasLimit,
		GasPrice: GasPrice,
		Value:    big.NewInt(0),
	})

	signedTx, err := etypes.SignTx(etx, signer, senderKey.PrivateKey())
	if err != nil {
		return nil, err
	}
	bs, err := signedTx.MarshalBinary()
	if err != nil {
		return nil, err
	}

	var to common.Address
	if c.Address != nil {
		to = *c.Address
	}

	return &TxWithSigner{
		Tx: &types.Transaction{
			ChainID:     chainID,
			Version:     Version,
			Timestamp:   uint64(time.Now().UnixNano()),
			Seq:         nonce,
			To:          to,
			Method:      "",
			GasPrice:    GasPrice,
			UseSeq:      true,
			IsEtherType: true,
			VmType:      types.Evm,
			Args:        bs,
		},
		Signer: senderKey,
	}, nil
}

// GetAbi returns the artifact and abi of Evm contract
func GetAbi(abiPath string) (map[string]interface{}, *abi.ABI, error) {

	path, _ := os.Getwd()
	b, err := os.ReadFile(path + abiPath)
	if err != nil {
		return nil, nil, ErrFileRead
	}
	var artifact map[string]interface{}
	if err := json.Unmarshal(b, &artifact); err != nil {
		return nil, nil, err
	}

	abiBytes, err := json.Marshal(artifact["abi"])
	if err != nil {
		return nil, nil, err
	}
	abi, err := abi.JSON(bytes.NewBuffer(abiBytes))
	if err != nil {
		return nil, nil, err
	}

	return artifact, &abi, nil
}

// NewEvmContract makes new Contract, abiPath is relative path
func NewEvmContractTx(cont IEvmContract, senderKey key.Key, nonceIncrement uint64, provider types.Provider, abiPath string, args ...any) (*TxWithSigner, error) {

	artifact, abi, err := GetAbi(abiPath)
	if err != nil {
		return nil, err
	}

	cont.Set(abi, provider, nil)
	bytecode, err := hex.DecodeString(strings.ReplaceAll(artifact["bytecode"].(string), "0x", ""))
	if err != nil {
		return nil, err
	}
	// data, err := abi.Pack("", args)  : 에러발생 : bug
	data, err := pack.Pack(args)
	if err != nil {
		return nil, err
	}

	data = append(bytecode, data...)
	tx, err := cont.makeEvmTx(senderKey, nonceIncrement, data)

	return tx, err
}

// type RewardPool Contract
// source code : fleta2.0/contracts/RewardPool/RewardPool.sol
type RewardPool struct {
	EvmContract
}

// type UserReward
type UserReward struct {
	User   common.Address
	Amount *big.Int
}

// NewRewardPoolTx makes New Reward Pool deploy Transaction
func NewRewardPoolTx(senderKey key.Key, nonceIncrement uint64, provider types.Provider, tokenAddress *common.Address) (*RewardPool, *TxWithSigner, error) {
	pool := &RewardPool{}
	tx, err := NewEvmContractTx(pool, senderKey, nonceIncrement, provider, "/abi/RewardPool.json", tokenAddress)
	return pool, tx, err

}

// AddRewardTx returns the addReward method tx of the RewardPool contract
func (c *RewardPool) AddRewardTx(senderKey key.Key, nonceIncrement uint64, total *big.Int, userRewards []UserReward) (*TxWithSigner, error) {
	return c.MakeEvmTx(senderKey, nonceIncrement, "addReward", total, userRewards)
}

// ClaimTx returns the claim method tx of the RewardPool contract
func (c *RewardPool) ClaimTx(senderKey key.Key, nonceIncrement uint64) (*TxWithSigner, error) {
	return c.MakeEvmTx(senderKey, nonceIncrement, "claim")
}

// Mrc20Token is deployed by Go-Contract, but has a similiar ERC20-token-abi
type Mrc20Token struct {
	EvmContract
}

// NewMrc20Token deploy Go Token Contract and convert to Evm MRC20 Token
func NewMrc20Token(tb *TestBlockChain, senderKey key.Key, name, symbol string) (*Mrc20Token, error) {

	sender := senderKey.PublicKey().Address()
	tx, err := DeployTokenTx(tb, senderKey, name, symbol,
		map[common.Address]*amount.Amount{
			sender: amount.NewAmount(100000000, 0),
		})
	if err != nil {
		return nil, err
	}

	b := tb.MustAddBlock([]*TxWithSigner{tx})

	var contractAddress common.Address
	for _, event := range b.Body.Events {
		if event.Index == 0 && event.Type == ctypes.EventTagTxMsg {
			contractAddress = common.BytesToAddress(event.Result)
			log.Println(symbol, "address = ", contractAddress)
		}
	}

	token, err := NewMrc20TokenFromAddress(&contractAddress, tb.Provider)

	return token, nil
}

// NewMrc20TokenFromAddress makes Mrc20Token EVm Contract with address
func NewMrc20TokenFromAddress(contractAddress *common.Address, provider types.Provider) (*Mrc20Token, error) {
	_, abi, err := GetAbi("/abi/IMRC20.json")
	if err != nil {
		return nil, err
	}
	token := &Mrc20Token{}
	token.Set(abi, provider, contractAddress)
	return token, nil
}

// ApproveTx returns the approve method tx of the Mrc20 contract
func (c *Mrc20Token) ApproveTx(senderKey key.Key, nonceIncrement uint64, to *common.Address, amount *big.Int) (*TxWithSigner, error) {
	return c.MakeEvmTx(senderKey, nonceIncrement, "approve", to, amount)
}
