package nft721receiver

import (
	"bytes"
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/core/types"
)

type NFT721ReceiverContractConstruction struct {
}

func (s *NFT721ReceiverContractConstruction) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

func (s *NFT721ReceiverContractConstruction) ReadFrom(r io.Reader) (int64, error) {
	return 0, nil
}

type NFT721ReceiverContract struct {
	addr   common.Address
	master common.Address
}

// func (cont *NFT721ReceiverContract) Name() string {
// 	return "NFT721"
// }

func (cont *NFT721ReceiverContract) Address() common.Address {
	return cont.addr
}

func (cont *NFT721ReceiverContract) Master() common.Address {
	return cont.master
}

func (cont *NFT721ReceiverContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

func (cont *NFT721ReceiverContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &NFT721ReceiverContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}
	return nil
}

func (cont *NFT721ReceiverContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

//////////////////////////////////////////////////
// Public Read Functions
//////////////////////////////////////////////////
func SupportsInterface(interfaceID []byte) bool {
	switch 0 {
	case bytes.Compare(interfaceID, []byte{0x15, 0x0b, 0x7a, 0x02}):
		return true
	}
	return false
}

//////////////////////////////////////////////////
// Public Read Functions
//////////////////////////////////////////////////

/// @notice A descriptive name for a collection of NFTs in this contract
func name(cc *types.ContractContext) string {
	return "REceiver"
}
