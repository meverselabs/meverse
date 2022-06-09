package nft721receiver

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
)

func (cont *NFT721ReceiverContract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *NFT721ReceiverContract
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

/// @notice A descriptive name for a collection of NFTs in this contract
func (f *front) Name(cc *types.ContractContext) string {
	return name(cc)
}

/// @notice A descriptive name for a collection of NFTs in this contract
func (f *front) OnERC721Received(cc *types.ContractContext, _operator common.Address, _from common.Address, _tokenId hash.Hash256, _data []byte) []byte {
	return OnERC721Received(_operator, _from, _tokenId, _data)
}
