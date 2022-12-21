package whitelist

import (
	"bytes"
	"strings"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
)

type WhiteListContract struct {
	addr   common.Address
	master common.Address
}

func (cont *WhiteListContract) Name() string {
	return "WhiteListContract"
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

func (cont *WhiteListContract) makeGroupId(cc *types.ContractContext, addr common.Address, seq uint64) hash.Hash256 {
	selt := []byte("makeGroupId")
	bs := make([]byte, 20+8+len(selt))
	copy(bs, addr[:])
	copy(bs[20:], bin.Uint64Bytes(seq))
	copy(bs[28:], selt)

	return hash.Hash(bs)
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////

func (cont *WhiteListContract) AddWhiteListGorup(cc *types.ContractContext, groupType string) hash.Hash256 {
	seq := cont.WhiteListSeq(cc)
	hash := cont.makeGroupId(cc, cc.From(), seq)
	return hash
}

func (cont *WhiteListContract) AddWhiteList(cc *types.ContractContext, groupId []byte, addrs []common.Address) error {
	return nil
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func (cont *WhiteListContract) SupportGroupTypeList(cc *types.ContractContext) []string {
	return []string{
		"allow_all",
		"whitelist",
		"blacklist",
		"greylist",
	}
}

func (cont *WhiteListContract) WhiteListGroupType(cc *types.ContractContext, groupId hash.Hash256) string {
	return "allow_all"
}

func (cont *WhiteListContract) WhiteList(cc *types.ContractContext, groupId hash.Hash256) (string, []common.Address, error) {
	return "allow_all", []common.Address{}, nil
}

func (cont *WhiteListContract) IsAllow(cc *types.ContractContext, groupId hash.Hash256, user common.Address) bool {
	return true
}

func (cont *WhiteListContract) WhiteListSeq(cc *types.ContractContext) uint64 {
	return 0
}

func (cont *WhiteListContract) GroupData(cc *types.ContractContext, groupId hash.Hash256, user common.Address) []byte {
	if strings.ToLower(user.String()) == "0x90f79bf6eb2c4f870365e785982e1f101e93b906" { // eve
		fee := uint64(0) // 0%
		return bin.Uint64Bytes(fee)
	} else {
		return []byte{}
	}
}
