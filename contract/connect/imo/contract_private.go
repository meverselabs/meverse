package imo

import (
	"bytes"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/types"
)

//////////////////////////////////////////////////
// private Write Functions
//////////////////////////////////////////////////
func (cont *ImoContract) setUserInfo(cc *types.ContractContext, user common.Address, userInfo *UserInfo) {
	bf := new(bytes.Buffer)
	userInfo.WriteTo(bf)
	cc.SetContractData(makeUserInfoKey(user), bf.Bytes())
}

func (cont *ImoContract) addAddress(cc *types.ContractContext, user common.Address) error {
	al, err := cont.AddressList(cc)
	if err != nil {
		return err
	}
	al = append(al, user)

	bf := new(bytes.Buffer)
	if _, err := bin.WriteUint16(bf, uint16(len(al))); err != nil {
		return err
	} else {
		for i := 0; i < len(al); i++ {
			if _, err := bin.WriteBytes(bf, al[i][:]); err != nil {
				return err
			}
		}
	}
	cc.SetContractData([]byte{tagAddressList}, bf.Bytes())
	return nil
}

func (cont *ImoContract) addTotalAmount(cc *types.ContractContext, amt *amount.Amount) {
	totalAmt := cont.TotalAmount(cc)
	cc.SetContractData([]byte{tagTotalAmount}, totalAmt.Add(amt).Bytes())
}

func (cont *ImoContract) _userInfo(cc *types.ContractContext, user common.Address) (*UserInfo, error) {
	bs := cc.ContractData(makeUserInfoKey(user))

	if len(bs) != 0 {
		data := &UserInfo{}
		if _, err := data.ReadFrom(bytes.NewReader(bs)); err != nil {
			return nil, err
		}
		return data, nil
	} else {
		return &UserInfo{
			Amt:     amount.NewAmount(0, 0),
			Claimed: false,
		}, nil
	}
}
