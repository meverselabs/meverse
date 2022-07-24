package imo

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
)

type ImoContract struct {
	addr   common.Address
	master common.Address
}

func (cont *ImoContract) Name() string {
	return "ImoContract"
}

func (cont *ImoContract) Address() common.Address {
	return cont.addr
}

func (cont *ImoContract) Master() common.Address {
	return cont.master
}

func (cont *ImoContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

func (cont *ImoContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &ImoContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}

	cc.SetContractData([]byte{tagProjectOwner}, data.ProjectOwner[:])
	cc.SetContractData([]byte{tagPayToken}, data.PayToken[:])
	cc.SetContractData([]byte{tagProjectToken}, data.ProjectToken.Bytes())
	cc.SetContractData([]byte{tagProjectOffering}, data.ProjectOffering.Bytes())
	cc.SetContractData([]byte{tagProjectRaising}, data.ProjectRaising.Bytes())
	cc.SetContractData([]byte{tagPayLimit}, data.PayLimit.Bytes())
	cc.SetContractData([]byte{tagStartBlock}, bin.Uint32Bytes(data.StartBlock))
	cc.SetContractData([]byte{tagEndBlock}, bin.Uint32Bytes(data.EndBlock))
	cc.SetContractData([]byte{tagHarvestFeeFactor}, bin.Uint16Bytes(data.HarvestFeeFactor))
	cc.SetContractData([]byte{tagWhiteListAddress}, data.WhiteListAddress[:])
	cc.SetContractData([]byte{tagWhiteListGroupId}, data.WhiteListGroupId[:])
	return nil
}

func (cont *ImoContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func (cont *ImoContract) ProjectOwner(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagProjectOwner})
	return common.BytesToAddress(bs)
}
func (cont *ImoContract) IsOwner(cc *types.ContractContext) bool {
	return cont.ProjectOwner(cc) == cc.From()
}

func (cont *ImoContract) PayToken(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagPayToken})
	return common.BytesToAddress(bs)
}
func (cont *ImoContract) ProjectToken(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagProjectToken})
	return common.BytesToAddress(bs)
}
func (cont *ImoContract) ProjectOffering(cc *types.ContractContext) *amount.Amount {
	bs := cc.ContractData([]byte{tagProjectOffering})
	return amount.NewAmountFromBytes(bs)
}
func (cont *ImoContract) ProjectRaising(cc *types.ContractContext) *amount.Amount {
	bs := cc.ContractData([]byte{tagProjectRaising})
	return amount.NewAmountFromBytes(bs)
}
func (cont *ImoContract) PayLimit(cc *types.ContractContext) *amount.Amount {
	bs := cc.ContractData([]byte{tagPayLimit})
	return amount.NewAmountFromBytes(bs)
}
func (cont *ImoContract) StartBlock(cc *types.ContractContext) uint32 {
	bs := cc.ContractData([]byte{tagStartBlock})
	return bin.Uint32(bs)
}
func (cont *ImoContract) EndBlock(cc *types.ContractContext) uint32 {
	bs := cc.ContractData([]byte{tagEndBlock})
	return bin.Uint32(bs)
}
func (cont *ImoContract) HarvestFeeFactor(cc *types.ContractContext) uint16 {
	bs := cc.ContractData([]byte{tagHarvestFeeFactor})
	return bin.Uint16(bs)
}
func (cont *ImoContract) WhiteListAddress(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagWhiteListAddress})
	return common.BytesToAddress(bs)
}
func (cont *ImoContract) WhiteListGroupId(cc *types.ContractContext) hash.Hash256 {
	bs := cc.ContractData([]byte{tagWhiteListGroupId})
	var hs hash.Hash256
	copy(hs[:], bs)
	return hs
}
func (cont *ImoContract) TotalAmount(cc *types.ContractContext) *amount.Amount {
	bs := cc.ContractData([]byte{tagTotalAmount})
	return amount.NewAmountFromBytes(bs)
}

func (cont *ImoContract) UserInfo(cc *types.ContractContext, user common.Address) (*amount.Amount, bool, error) {
	us, err := cont._userInfo(cc, user)
	if err != nil {
		return nil, false, err
	}
	return us.Amt, us.Claimed, nil
}

func (cont *ImoContract) AddressList(cc *types.ContractContext) ([]common.Address, error) {
	bs := cc.ContractData([]byte{tagAddressList})
	if len(bs) == 0 {
		return []common.Address{}, nil
	}
	bf := bytes.NewBuffer(bs)
	if Len, _, err := bin.ReadUint16(bf); err != nil {
		return nil, err
	} else {
		p := make([]common.Address, Len)
		for i := 0; i < int(Len); i++ {
			if addrbs, _, err := bin.ReadBytes(bf); err != nil {
				return nil, err
			} else {
				copy(p[i][:], addrbs)
			}
		}
		return p, nil
	}
}

func (cont *ImoContract) HasHarvest(cc *types.ContractContext, _user common.Address) (bool, error) {
	if us, err := cont._userInfo(cc, _user); err != nil {
		return false, err
	} else {
		return us.Claimed, nil
	}
}

// 1e18 == 100%
func (cont *ImoContract) GetUserAllocation(cc *types.ContractContext, _user common.Address) *big.Int {
	if us, err := cont._userInfo(cc, _user); err != nil {
		return big.NewInt(0)
	} else {
		totalAmount := cont.TotalAmount(cc)
		if !totalAmount.IsZero() {
			return us.Amt.Div(totalAmount).Int
		}
	}
	return big.NewInt(0)
}

// get the amount of IFO token you will get
func (cont *ImoContract) GetOfferingAmount(cc *types.ContractContext, _user common.Address) (*amount.Amount, error) {
	totalAmount := cont.TotalAmount(cc)
	projectRaising := cont.ProjectRaising(cc)

	projectoffering := cont.ProjectOffering(cc)
	if totalAmount.Cmp(projectRaising.Int) > 0 {
		ballocation := cont.GetUserAllocation(cc, _user)
		allocation := &amount.Amount{Int: ballocation}
		return projectoffering.Mul(allocation), nil
	} else {
		ui, err := cont._userInfo(cc, _user)
		if err != nil {
			return nil, err
		}
		if !projectRaising.IsZero() {
			return ui.Amt.Mul(projectoffering).Div(projectRaising), nil
		}
	}
	return amount.NewAmount(0, 0), nil
}

// get the amount of lp token you will be refunded
func (cont *ImoContract) GetRefundingAmount(cc *types.ContractContext, _user common.Address) (*amount.Amount, error) {
	totalAmount := cont.TotalAmount(cc)
	projectRaising := cont.ProjectRaising(cc)

	projectoffering := cont.ProjectOffering(cc)
	if totalAmount.Cmp(projectRaising.Int) <= 0 {
		return amount.NewAmount(0, 0), nil
	}
	ballocation := cont.GetUserAllocation(cc, _user)
	allocation := &amount.Amount{Int: ballocation}
	payAmount := projectoffering.Mul(allocation)
	ui, err := cont._userInfo(cc, _user)
	if err != nil {
		return nil, err
	}
	return ui.Amt.Sub(payAmount), nil
}

func (cont *ImoContract) GetAddressListLength(cc *types.ContractContext) (int, error) {
	addrList, err := cont.AddressList(cc)
	if err != nil {
		return 0, err
	}
	return len(addrList), nil
}

func (cont *ImoContract) CheckWhiteList(cc *types.ContractContext, user common.Address) (bool, error) {
	addr := cont.WhiteListAddress(cc)
	if common.ZeroAddr == addr {
		return true, nil
	}
	groupId := cont.WhiteListGroupId(cc)
	if ins, err := cc.Exec(cc, addr, "IsAllow", []interface{}{groupId, cc.From()}); err != nil {
		return false, err
	} else if len(ins) == 0 {
		return false, errors.New("invalid " + addr.String() + " IsAllow")
	} else if val, ok := ins[0].(bool); !ok {
		return false, errors.New("invalid " + addr.String() + " IsAllow !bool")
	} else {
		return val, nil
	}
}
