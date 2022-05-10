package nft1155

import (
	"bytes"
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
)

type NFT1155Contract struct {
	addr   common.Address
	master common.Address
}

func (cont *NFT1155Contract) Name() string {
	return "NFT1155"
}

func (cont *NFT1155Contract) Address() common.Address {
	return cont.addr
}

func (cont *NFT1155Contract) Master() common.Address {
	return cont.master
}

func (cont *NFT1155Contract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

func (cont *NFT1155Contract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &NFT1155ContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}

	cc.SetContractData([]byte{tagName}, []byte(data.Name))
	cc.SetContractData([]byte{tagSymbol}, []byte(data.Symbol))
	return nil
}

func (cont *NFT1155Contract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

//////////////////////////////////////////////////
// Public Read Functions
//////////////////////////////////////////////////
// bytes4 constant public ERC1155_ERC165 = 0xd9b67a26; // ERC-165 identifier for the main token standard.
// bytes4 constant public ERC1155_ERC165_TOKENRECEIVER = 0x4e2312e0; // ERC-165 identifier for the `ERC1155TokenReceiver` support (i.e. `bytes4(keccak256("onERC1155Received(address,address,uint256,uint256,bytes)")) ^ bytes4(keccak256("onERC1155BatchReceived(address,address,uint256[],uint256[],bytes)"))`).
// bytes4 constant public ERC1155_ACCEPTED = 0xf23a6e61; // Return value from `onERC1155Received` call if a contract accepts receipt (i.e `bytes4(keccak256("onERC1155Received(address,address,uint256,uint256,bytes)"))`).
// bytes4 constant public ERC1155_BATCH_ACCEPTED = 0xbc197c81; // Return value from `onERC1155BatchReceived` call if a contract accepts receipt (i.e `bytes4(keccak256("onERC1155BatchReceived(address,address,uint256[],uint256[],bytes)"))`).

func (cont *NFT1155Contract) supportsInterface(interfaceID []byte) bool {
	switch 0 {
	case bytes.Compare(interfaceID, []byte{0x01, 0xff, 0xc9, 0xa7}):
		return true
	}
	return false
}

/**
@notice Get the balance of an account's tokens.
@param _owner  The address of the token holder
@param _id     ID of the token
@return        The _owner's balance of the token type requested
*/
func (cont *NFT1155Contract) BalanceOf(_owner common.Address, _id hash.Hash256) *amount.Amount {
	return nil
}

/**
@notice Get the balance of multiple account/token pairs
@param _owners The addresses of the token holders
@param _ids    ID of the tokens
@return        The _owner's balance of the token types requested (i.e. balance for each (owner, id) pair)
*/
func (cont *NFT1155Contract) BalanceOfBatch(_owners []common.Address, _ids []hash.Hash256) []*amount.Amount {
	return nil
}

/**
@notice Queries the approval status of an operator for a given owner.
@param _owner     The owner of the tokens
@param _operator  Address of authorized operator
@return           True if the operator is approved, false if not
*/
func (cont *NFT1155Contract) IsApprovedForAll(_owner common.Address, _operator common.Address) bool {
	return false
}

/**
@notice A distinct Uniform Resource Identifier (URI) for a given token.
@dev URIs are defined in RFC 3986.
The URI MUST point to a JSON file that conforms to the "ERC-1155 Metadata URI JSON Schema".
@return URI string
*/
func (cont *NFT1155Contract) Uri(_id hash.Hash256) string {
	return ""
}

//////////////////////////////////////////////////
// Public Write Functions
//////////////////////////////////////////////////
/**
@notice Transfers `_value` amount of an `_id` from the `_from` address to the `_to` address specified (with safety call).
@dev Caller must be approved to manage the tokens being transferred out of the `_from` account (see "Approval" section of the standard).
MUST revert if `_to` is the zero address.
MUST revert if balance of holder for token `_id` is lower than the `_value` sent.
MUST revert on any other error.
MUST emit the `TransferSingle` event to reflect the balance change (see "Safe Transfer Rules" section of the standard).
After the above conditions are met, this function MUST check if `_to` is a smart contract (e.g. code size > 0). If so, it MUST call `onERC1155Received` on `_to` and act appropriately (see "Safe Transfer Rules" section of the standard).
@param _from    Source address
@param _to      Target address
@param _id      ID of the token type
@param _value   Transfer amount
@param _data    Additional data with no specified format, MUST be sent unaltered in call to `onERC1155Received` on `_to`
*/
func (cont *NFT1155Contract) SafeTransferFrom(_from common.Address, _to common.Address, _id hash.Hash256, _value *big.Int, _data []byte) {

}

/**
@notice Transfers `_values` amount(s) of `_ids` from the `_from` address to the `_to` address specified (with safety call).
@dev Caller must be approved to manage the tokens being transferred out of the `_from` account (see "Approval" section of the standard).
MUST revert if `_to` is the zero address.
MUST revert if length of `_ids` is not the same as length of `_values`.
MUST revert if any of the balance(s) of the holder(s) for token(s) in `_ids` is lower than the respective amount(s) in `_values` sent to the recipient.
MUST revert on any other error.
MUST emit `TransferSingle` or `TransferBatch` event(s) such that all the balance changes are reflected (see "Safe Transfer Rules" section of the standard).
Balance changes and events MUST follow the ordering of the arrays (_ids[0]/_values[0] before _ids[1]/_values[1], etc).
After the above conditions for the transfer(s) in the batch are met, this function MUST check if `_to` is a smart contract (e.g. code size > 0). If so, it MUST call the relevant `ERC1155TokenReceiver` hook(s) on `_to` and act appropriately (see "Safe Transfer Rules" section of the standard).
@param _from    Source address
@param _to      Target address
@param _ids     IDs of each token type (order and length must match _values array)
@param _values  Transfer amounts per token type (order and length must match _ids array)
@param _data    Additional data with no specified format, MUST be sent unaltered in call to the `ERC1155TokenReceiver` hook(s) on `_to`
*/
func (cont *NFT1155Contract) SafeBatchTransferFrom(_from common.Address, _to common.Address, _ids []hash.Hash256, _values []*big.Int, _data []byte) {
}

/**
@notice Enable or disable approval for a third party ("operator") to manage all of the caller's tokens.
@dev MUST emit the ApprovalForAll event on success.
@param _operator  Address to add to the set of authorized operators
@param _approved  True if the operator is approved, false to revoke approval
*/
func (cont *NFT1155Contract) SetApprovalForAll(_operator common.Address, _approved bool) {

}
