package erc20wrapper

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/ethereum/core/vm"
)

type Erc20WrapperContract struct {
	addr   common.Address
	master common.Address
}

func (cont *Erc20WrapperContract) Address() common.Address {
	return cont.addr
}

func (cont *Erc20WrapperContract) Master() common.Address {
	return cont.master
}

func (cont *Erc20WrapperContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

func (cont *Erc20WrapperContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &ERC20WrapperContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}

	cc.SetContractData([]byte{tagErc20Token}, data.Erc20Token[:])

	return nil
}

func (cont *Erc20WrapperContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////
// func (cont *Erc20WrapperContract) SetErc20Token(cc *types.ContractContext, erc20 common.Address) error {
// 	cc.SetContractData([]byte{tagErc20Token}, erc20.Bytes())
// 	return nil
// }

func (cont *Erc20WrapperContract) Approve(cc *types.ContractContext, spender common.Address, Amount *amount.Amount) error {
	method := "approve(address,uint256)"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	data, err := hex.DecodeString(id + addressToString32(spender) + fmt.Sprintf("%064x", Amount.Int))
	if err != nil {
		return err
	}
	erc20 := cont.Erc20Token(cc)
	_, _, err = cc.EvmCall(vm.AccountRef(cc.From()), erc20, data)
	return err
}

func (cont *Erc20WrapperContract) Transfer(cc *types.ContractContext, To common.Address, Amount *amount.Amount) error {
	method := "transfer(address,uint256)"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	data, err := hex.DecodeString(id + addressToString32(To) + fmt.Sprintf("%064x", Amount.Int))
	if err != nil {
		return err
	}
	erc20 := cont.Erc20Token(cc)
	_, _, err = cc.EvmCall(vm.AccountRef(cc.From()), erc20, data)
	return err
}

func (cont *Erc20WrapperContract) TransferFrom(cc *types.ContractContext, From common.Address, To common.Address, Amount *amount.Amount) error {
	method := "transferFrom(address,address,uint256)"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	data, err := hex.DecodeString(id + addressToString32(From) + addressToString32(To) + fmt.Sprintf("%064x", Amount.Int))
	if err != nil {
		return err
	}
	erc20 := cont.Erc20Token(cc)
	_, _, err = cc.EvmCall(vm.AccountRef(cc.From()), erc20, data)
	return err
}

func (cont *Erc20WrapperContract) IncreaseAllowance(cc *types.ContractContext, spender common.Address, addedValue *amount.Amount) error {
	method := "increaseAllowance(address,uint256)"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	data, err := hex.DecodeString(id + addressToString32(spender) + fmt.Sprintf("%064x", addedValue.Int))
	if err != nil {
		return err
	}

	erc20 := cont.Erc20Token(cc)
	_, _, err = cc.EvmCall(vm.AccountRef(cc.From()), erc20, data)
	return err
}

func (cont *Erc20WrapperContract) DecreaseAllowance(cc *types.ContractContext, spender common.Address, subtractedValue *amount.Amount) error {
	method := "decreaseAllowance(address,uint256)"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	data, err := hex.DecodeString(id + addressToString32(spender) + fmt.Sprintf("%064x", subtractedValue.Int))
	if err != nil {
		return err
	}
	erc20 := cont.Erc20Token(cc)
	_, _, err = cc.EvmCall(vm.AccountRef(cc.From()), erc20, data)
	return err
}

func (cont *Erc20WrapperContract) SetMinter(cc *types.ContractContext, To common.Address, Is bool) error {
	method := "setMinter(address,bool)"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	var IsVar uint
	if Is {
		IsVar = 1
	}
	data, err := hex.DecodeString(id + addressToString32(To) + fmt.Sprintf("%064x", IsVar))
	if err != nil {
		return err
	}
	erc20 := cont.Erc20Token(cc)
	_, _, err = cc.EvmCall(vm.AccountRef(cc.From()), erc20, data)
	return err
}

func (cont *Erc20WrapperContract) Mint(cc *types.ContractContext, To common.Address, Amount *amount.Amount) error {
	method := "mint(address,uint256)"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	data, err := hex.DecodeString(id + addressToString32(To) + fmt.Sprintf("%064x", Amount.Int))
	if err != nil {
		return err
	}
	erc20 := cont.Erc20Token(cc)
	_, _, err = cc.EvmCall(vm.AccountRef(cc.From()), erc20, data)
	return err
}

func (cont *Erc20WrapperContract) Burn(cc *types.ContractContext, Amount *amount.Amount) error {
	method := "burn(uint256)"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	data, err := hex.DecodeString(id + fmt.Sprintf("%064x", Amount.Int))
	if err != nil {
		return err
	}
	erc20 := cont.Erc20Token(cc)
	_, _, err = cc.EvmCall(vm.AccountRef(cc.From()), erc20, data)
	return err
}

func (cont *Erc20WrapperContract) BurnFrom(cc *types.ContractContext, Addr common.Address, Amount *amount.Amount) error {
	method := "burnFrom(address,uint256)"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	data, err := hex.DecodeString(id + addressToString32(Addr) + fmt.Sprintf("%064x", Amount.Int))
	if err != nil {
		return err
	}
	erc20 := cont.Erc20Token(cc)
	_, _, err = cc.EvmCall(vm.AccountRef(cc.From()), erc20, data)
	return err
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////
func (cont *Erc20WrapperContract) Erc20Token(cc types.ContractLoader) common.Address {
	return common.BytesToAddress(cc.ContractData([]byte{tagErc20Token}))
}

func (cont *Erc20WrapperContract) Name(cc *types.ContractContext) (string, error) {
	method := "name()"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	data, err := hex.DecodeString(id)
	if err != nil {
		return "", err
	}
	erc20 := cont.Erc20Token(cc)
	if ret, _, err := cc.EvmCall(vm.AccountRef(cc.From()), erc20, data); err != nil {
		return "", err
	} else {
		return unpackString(ret)
	}
}

func (cont *Erc20WrapperContract) Symbol(cc *types.ContractContext) (string, error) {
	method := "symbol()"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	data, err := hex.DecodeString(id)
	if err != nil {
		return "", err
	}
	erc20 := cont.Erc20Token(cc)
	if ret, _, err := cc.EvmCall(vm.AccountRef(cc.From()), erc20, data); err != nil {
		return "", err
	} else {
		return unpackString(ret)
	}
}

func (cont *Erc20WrapperContract) TotalSupply(cc *types.ContractContext) (*amount.Amount, error) {
	method := "totalSupply()"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	data, err := hex.DecodeString(id)
	if err != nil {
		return nil, err
	}
	erc20 := cont.Erc20Token(cc)
	if ret, _, err := cc.EvmCall(vm.AccountRef(cc.From()), erc20, data); err != nil {
		return nil, err
	} else {
		return toAmount(ret), nil
	}
}

func (cont *Erc20WrapperContract) Decimals(cc *types.ContractContext) (*big.Int, error) {
	method := "decimals()"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	data, err := hex.DecodeString(id)
	if err != nil {
		return nil, err
	}
	erc20 := cont.Erc20Token(cc)
	if ret, _, err := cc.EvmCall(vm.AccountRef(cc.From()), erc20, data); err != nil {
		return nil, err
	} else {
		return new(big.Int).SetBytes(ret), nil
	}
}

func (cont *Erc20WrapperContract) BalanceOf(cc *types.ContractContext, from common.Address) (*amount.Amount, error) {
	method := "balanceOf(address)"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	data, err := hex.DecodeString(id + addressToString32(from))
	if err != nil {
		return nil, err
	}
	erc20 := cont.Erc20Token(cc)
	if ret, _, err := cc.EvmCall(vm.AccountRef(cc.From()), erc20, data); err != nil {
		return nil, err
	} else {
		return toAmount(ret), nil
	}
}

func (cont *Erc20WrapperContract) IsMinter(cc *types.ContractContext, addr common.Address) (bool, error) {
	method := "isMinter(address)"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	data, err := hex.DecodeString(id + addressToString32(addr))
	if err != nil {
		return false, err
	}
	erc20 := cont.Erc20Token(cc)
	if ret, _, err := cc.EvmCall(vm.AccountRef(cc.From()), erc20, data); err != nil {
		return false, err
	} else {
		return new(big.Int).SetBytes(ret).Cmp(new(big.Int)) != 0, nil
	}
}

func (cont *Erc20WrapperContract) Allowance(cc *types.ContractContext, _owner common.Address, _spender common.Address) (*amount.Amount, error) {
	method := "allowance(address,address)"
	id := hex.EncodeToString(crypto.Keccak256([]byte(method))[:4])

	data, err := hex.DecodeString(id + addressToString32(_owner) + addressToString32(_spender))
	if err != nil {
		return nil, err
	}
	erc20 := cont.Erc20Token(cc)
	if ret, _, err := cc.EvmCall(vm.AccountRef(cc.From()), erc20, data); err != nil {
		return nil, err
	} else {
		return toAmount(ret), nil
	}
}
