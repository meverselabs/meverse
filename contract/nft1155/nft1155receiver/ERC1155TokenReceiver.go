package nftreceiver

import (
	"bytes"
	"encoding/hex"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
)

// bytes4 constant public ERC1155_ERC165_TOKENRECEIVER = 0x4e2312e0; // ERC-165 identifier for the `ERC1155TokenReceiver` support (i.e. `bytes4(keccak256("onERC1155Received(address,address,uint256,uint256,bytes)")) ^ bytes4(keccak256("onERC1155BatchReceived(address,address,uint256[],uint256[],bytes)"))`).
// bytes4 constant public ERC1155_ACCEPTED = 0xf23a6e61; // Return value from `onERC1155Received` call if a contract accepts receipt (i.e `bytes4(keccak256("onERC1155Received(address,address,uint256,uint256,bytes)"))`).
// bytes4 constant public ERC1155_BATCH_ACCEPTED = 0xbc197c81; // Return value from `onERC1155BatchReceived` call if a contract accepts receipt (i.e `bytes4(keccak256("onERC1155BatchReceived(address,address,uint256[],uint256[],bytes)"))`).

func SupportsInterface(interfaceID []byte) bool {
	switch 0 {
	case bytes.Compare(interfaceID, []byte{0x4e, 0x23, 0x12, 0xe0}), // ERC1155_ERC165_TOKENRECEIVER
		bytes.Compare(interfaceID, []byte{0xf2, 0x3a, 0x6e, 0x61}), // ERC1155_ACCEPTED
		bytes.Compare(interfaceID, []byte{0xbc, 0x19, 0x7c, 0x81}): // ERC1155_BATCH_ACCEPTED
		return true
	}
	return false
}

/**
  @notice Handle the receipt of a single ERC1155 token type.
  @dev An ERC1155-compliant smart contract MUST call this function on the token recipient contract, at the end of a `safeTransferFrom` after the balance has been updated.
  This function MUST return `bytes4(keccak256("onERC1155Received(address,address,uint256,uint256,bytes)"))` (i.e. 0xf23a6e61) if it accepts the transfer.
  This function MUST revert if it rejects the transfer.
  Return of any other value than the prescribed keccak256 generated value MUST result in the transaction being reverted by the caller.
  @param _operator  The address which initiated the transfer (i.e. msg.sender)
  @param _from      The address which previously owned the token
  @param _id        The ID of the token being transferred
  @param _value     The amount of tokens being transferred
  @param _data      Additional data with no specified format
  @return           `bytes4(keccak256("onERC1155Received(address,address,uint256,uint256,bytes)"))`
*/
func OnERC1155Received(_operator common.Address, _from common.Address, _id hash.Hash256, _value hash.Hash256, _data []byte) ([]byte, error) {
	return hex.DecodeString("0xf23a6e61")
}

/**
  @notice Handle the receipt of multiple ERC1155 token types.
  @dev An ERC1155-compliant smart contract MUST call this function on the token recipient contract, at the end of a `safeBatchTransferFrom` after the balances have been updated.
  This function MUST return `bytes4(keccak256("onERC1155BatchReceived(address,address,uint256[],uint256[],bytes)"))` (i.e. 0xbc197c81) if it accepts the transfer(s).
  This function MUST revert if it rejects the transfer(s).
  Return of any other value than the prescribed keccak256 generated value MUST result in the transaction being reverted by the caller.
  @param _operator  The address which initiated the batch transfer (i.e. msg.sender)
  @param _from      The address which previously owned the token
  @param _ids       An array containing ids of each token being transferred (order and length must match _values array)
  @param _values    An array containing amounts of each token being transferred (order and length must match _ids array)
  @param _data      Additional data with no specified format
  @return           `bytes4(keccak256("onERC1155BatchReceived(address,address,uint256[],uint256[],bytes)"))`
*/
func OnERC1155BatchReceived(_operator common.Address, _from common.Address, _ids []hash.Hash256, _values []hash.Hash256, _data []byte) ([]byte, error) {
	return hex.DecodeString("0xbc197c81")

}
