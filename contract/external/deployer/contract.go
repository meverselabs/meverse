package deployer

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/contract/external/engin/engincontext"
	"github.com/meverselabs/meverse/core/types"
)

type DeployerContract struct {
	addr   common.Address
	master common.Address
}

func (cont *DeployerContract) Name() string {
	return "DeployerContract"
}

func (cont *DeployerContract) Address() common.Address {
	return cont.addr
}

func (cont *DeployerContract) Master() common.Address {
	return cont.master
}

func (cont *DeployerContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

func (cont *DeployerContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &DeployerContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}

	cc.SetContractData([]byte{tagEnginAddress}, data.EnginAddress[:])
	cc.SetContractData([]byte{tagEnginName}, []byte(data.EnginName))
	cc.SetContractData([]byte{tagEnginVersion}, bin.Uint32Bytes(data.EnginVersion))
	cc.SetContractData([]byte{tagBin}, data.Binary)
	cc.SetContractData([]byte{tagOwner}, data.Owner[:])
	if data.Updateable {
		cc.SetContractData([]byte{tagUpdateable}, []byte{1})
	} else {
		cc.SetContractData([]byte{tagUpdateable}, []byte{})
	}
	return nil
}

func (cont *DeployerContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

//////////////////////////////////////////////////
// Private Functions
//////////////////////////////////////////////////
func isOwner(cc *types.ContractContext) bool {
	ownerBs := cc.ContractData([]byte{tagOwner})
	var Owner common.Address
	copy(Owner[:], ownerBs)
	return Owner == cc.From()
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////

func (cont *DeployerContract) SetOwner(cc *types.ContractContext, NewOwner common.Address) error {
	if !isOwner(cc) {
		return errors.New("is not owner")
	}

	ownerBs := cc.ContractData([]byte{tagOwner})
	var Owner common.Address
	copy(Owner[:], ownerBs)
	if Owner == NewOwner {
		return errors.New("new owner equare old owner")
	}

	cc.SetContractData([]byte{tagOwner}, NewOwner[:])
	return nil
}

func (cont *DeployerContract) Update(cc *types.ContractContext, EnginName string, EnginVersion uint32, contract []byte) error {
	if !isOwner(cc) {
		return errors.New("is not owner")
	}

	if !isUpdateable(cc) {
		return errors.New("is not updateable contract")
	}

	if EnginName != "" {
		cc.SetContractData([]byte{tagEnginName}, []byte(EnginName))
	}
	if EnginVersion != 0 {
		cc.SetContractData([]byte{tagEnginVersion}, bin.Uint32Bytes(EnginVersion))
	}
	cc.SetContractData([]byte{tagBin}, contract)
	eg, ecc, err := cont.getEngin(cc)
	if err != nil {
		return err
	}
	return eg.UpdateContract(ecc, contract)
}

func (cont *DeployerContract) InitContract(cc *types.ContractContext, contract []byte, params []interface{}) error {
	eg, ecc, err := cont.getEngin(cc)
	if err != nil {
		return err
	}
	return eg.InitContract(ecc, contract, params)
}

func (cont *DeployerContract) ContractInvoke(cc *types.ContractContext, method string, params []interface{}) (interface{}, error) {
	eg, ecc, err := cont.getEngin(cc)
	if err != nil {
		return nil, err
	}
	for i, v := range params {
		if am, ok := v.(*amount.Amount); ok {
			params[i] = am.Int
		}
	}
	res, err := eg.ContractInvoke(ecc, method, params)
	if err != nil {
		fmt.Printf("%+v", err)
	}
	return res, err
}

func (cont *DeployerContract) getEngin(cc *types.ContractContext) (types.IEngin, *engincontext.EnginContextContract, error) {
	EnginAddressBs := cc.ContractData([]byte{tagEnginAddress})
	EnginAddress := common.BytesToAddress(EnginAddressBs)
	EnginNameBs := cc.ContractData([]byte{tagEnginName})
	EnginName := string(EnginNameBs)
	EnginVersionBs := cc.ContractData([]byte{tagEnginVersion})
	EnginVersion := bin.Uint32(EnginVersionBs)

	var eg types.IEngin
	var ecc *engincontext.EnginContextContract
	if iss, err := cc.Exec(cc, EnginAddress, "LoadEngin", []interface{}{EnginName, EnginVersion}); err != nil {
		return nil, nil, err
	} else if len(iss) == 0 {
		return nil, nil, errors.New("error engin load")
	} else if _eg, ok := iss[0].(types.IEngin); !ok {
		return nil, nil, errors.New("invalid engin functions")
	} else {
		ecc = engincontext.NewEnginContextContract(cont.addr, cont.master, cc)
		eg = _eg
	}
	return eg, ecc, nil
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func isUpdateable(cc *types.ContractContext) bool {
	updateable := cc.ContractData([]byte{tagUpdateable})
	if len(updateable) == 0 || updateable[0] != 1 {
		return false
	}
	return true
}
