package nft721receiver

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
)

/// @notice Handle the receipt of an NFT
/// @dev The ERC721 smart contract calls this function on the recipient
///  after a `transfer`. This function MAY throw to revert and reject the
///  transfer. Return of other than the magic value MUST result in the
///  transaction being reverted.
///  Note: the contract address is always the message sender.
/// @param _operator The address which called `safeTransferFrom` function
/// @param _from The address which previously owned the token
/// @param _tokenId The NFT identifier which is being transferred
/// @param _data Additional data with no specified format
/// @return `bytes4(keccak256("onERC721Received(address,address,uint256,bytes)"))`
///  unless throwing
func OnERC721Received(_operator common.Address, _from common.Address, _tokenId hash.Hash256, _data []byte) []byte {
	return []byte{0x15, 0x0b, 0x7a, 0x02}
}
