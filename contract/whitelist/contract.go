package whitelist

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
)

type WhiteListContract struct {
	addr   common.Address
	master common.Address
}

func (cont *WhiteListContract) Name() string {
	return "WhiteList"
}

func (cont *WhiteListContract) Address() common.Address {
	return cont.addr
}

func (cont *WhiteListContract) Master() common.Address {
	return cont.master
}

func (cont *WhiteListContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

func (cont *WhiteListContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &WhiteListContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}

	// cc.SetContractData([]byte{tagStartBlock}, bin.Uint32Bytes(data.StartBlock))
	return nil
}

func (cont *WhiteListContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

//////////////////////////////////////////////////
// Private Writer Functions
//////////////////////////////////////////////////

func increaseWhiteListSeq(cc *types.ContractContext) *big.Int {
	bs := cc.ContractData([]byte{tagSeq})
	seq := big.NewInt(0).SetBytes(bs)
	seq.Add(seq, big.NewInt(1))
	cc.SetContractData([]byte{tagSeq}, seq.Bytes())
	return seq
}

func makeGroupId(cc *types.ContractContext, owner common.Address) hash.Hash256 {
	seq := increaseWhiteListSeq(cc)
	bs := append(owner[:], seq.Bytes()...)
	bs = append(bs, []byte("WhiteListMakeGroupId")...)
	return hash.Hash(bs)
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////

func (cont *WhiteListContract) AddGroup(cc *types.ContractContext, delegate common.Address, method string, params []interface{}, checkResult string, result []byte) (hash.Hash256, error) {
	owner := cc.From()
	groupId := makeGroupId(cc, owner)

	setOwner(cc, groupId, owner)
	return groupId, updateGroupData(cc, groupId, delegate, method, params, checkResult, result)
}

func setOwner(cc *types.ContractContext, groupId hash.Hash256, owner common.Address) {
	cc.SetContractData(makeGroupOwnerKey(groupId), owner[:])
}

func updateGroupData(cc *types.ContractContext, groupId hash.Hash256, delegate common.Address, method string, params []interface{}, checkResult string, result []byte) error {
	gd := &GroupData{
		delegate:    delegate,
		method:      method,
		params:      params,
		checkResult: checkResult,
		result:      result,
	}
	bb := &bytes.Buffer{}
	_, err := gd.WriteTo(bb)
	if err != nil {
		return err
	}
	cc.SetContractData(makeGroupDataKey(groupId), bb.Bytes())
	return nil
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

type bytable interface {
	Bytes() []byte
}

func (cont *WhiteListContract) GroupData(cc *types.ContractContext, groupId hash.Hash256, user common.Address) ([]byte, error) {
	bs := cc.ContractData(makeGroupDataKey(groupId))
	if len(bs) == 0 {
		return []byte{}, nil
	}
	bb := bytes.NewBuffer(bs)
	gd := &GroupData{}
	_, err := gd.ReadFrom(bb)
	if err != nil {
		return nil, nil
	}

	m := map[string]interface{}{
		"user": user,
	}
	gd.ParseParam(m)

	if inf, err := cc.Exec(cc, gd.delegate, gd.method, gd.params); err != nil {
		return nil, err
	} else {
		ors := strings.Split(gd.checkResult, ":")
		if len(ors) == 0 {
			ors = []string{"[]byte"}
		}
		switch ors[0] {
		case "[]byte":
			if len(inf) == 0 {
				return nil, nil
			} else if bs, ok := inf[0].([]byte); ok {
				return bs, nil
			} else if bsi, ok := inf[0].(bytable); ok {
				return bsi.Bytes(), nil
			} else {
				return nil, fmt.Errorf("not parsable data %v", inf[0])
			}
		case "*big.Int":
			if len(inf) == 0 {
				return nil, errors.New("invalid result")
			} else if bi, ok := inf[0].(*big.Int); !ok {
				return nil, fmt.Errorf("not parsable data %v", inf[0])
			} else {
				if len(ors) == 1 {
					return bi.Bytes(), nil
				}
				switch ors[1] {
				case "has":
					if bi.Cmp(big.NewInt(0)) > 0 {
						return gd.result, nil
					} else {
						return nil, nil
					}
				default:
					return nil, fmt.Errorf("invalid parse %v", ors[1])
				}
			}
		}
		return nil, errors.New("")
	}

	// if user.String() == "0xf3dA6Ce653D680EBAcC26873d38F91aCf33C56Ac" { // eve
	// 	fee := uint64(0) // 0%
	// 	return bin.Uint64Bytes(fee), nil
	// } else {
	// 	return []byte{}, nil
	// }
}
