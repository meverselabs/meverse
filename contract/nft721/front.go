package nft721

import (
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/contract/nft721/nft721regacy"
	"github.com/meverselabs/meverse/core/types"
)

func (cont *NFT721Contract) Front() interface{} {
	return &front{
		cont: cont,
	}
}

type front struct {
	cont *NFT721Contract
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

/// @notice A descriptive name for a collection of NFTs in this contract
func (f *front) Name(cc *types.ContractContext) string {
	return name(cc)
}

/// @notice An abbreviated name for NFTs in this contract
func (f *front) Symbol(cc *types.ContractContext) string {
	return symbol(cc)
}

/// @notice A distinct Uniform Resource Identifier (URI) for a given asset.
/// @dev Throws if `_tokenId` is not a valid NFT. URIs are defined in RFC
///  3986. The URI may point to a JSON file that conforms to the "ERC721
///  Metadata JSON Schema".
func (f *front) TokenURI(cc *types.ContractContext, _tokenId hash.Hash256) string {
	return f.cont.tokenURI(cc, _tokenId)
}
func (f *front) BaseURI(cc *types.ContractContext) string {
	return f.cont.baseURI(cc)
}

/// @notice Count all NFTs assigned to an owner
/// @dev NFTs assigned to the zero address are considered invalid, and this
///  function throws for queries about the zero address.
/// @param _owner An address for whom to query the balance
/// @return The number of NFTs owned by `_owner`, possibly zero
func (f *front) BalanceOf(cc *types.ContractContext, _owner common.Address) *big.Int {
	return balanceOf(cc, _owner)
}

/// @notice Find the owner of an NFT
/// @dev NFTs assigned to zero address are considered invalid, and queries
///  about them do throw.
/// @param _tokenId The identifier for an NFT
/// @return The address of the owner of the NFT
func (f *front) OwnerOf(cc *types.ContractContext, _tokenId hash.Hash256) common.Address {
	return ownerOf(cc, _tokenId)
}

/// @notice Get the approved address for a single NFT
/// @dev Throws if `_tokenId` is not a valid NFT.
/// @param _tokenId The NFT to find the approved address for
/// @return The approved address for this NFT, or the zero address if there is none
func (f *front) GetApproved(cc *types.ContractContext, _tokenId hash.Hash256) common.Address {
	return getApproved(cc, _tokenId)
}

/// @notice Query if an address is an authorized operator for another address
/// @param _owner The address that owns the NFTs
/// @param _operator The address that acts on behalf of the owner
/// @return True if `_operator` is an approved operator for `_owner`, false otherwise
func (f *front) IsApprovedForAll(cc *types.ContractContext, _owner common.Address, _operator common.Address) bool {
	return isApprovedForAll(cc, _owner, _operator)
}

/// @notice Count NFTs tracked by this contract
/// @return A count of valid NFTs tracked by this contract, where each one of
///  them has an assigned and queryable owner not equal to the zero address
func (f *front) TotalSupply(cc *types.ContractContext) *big.Int {
	return totalSupply(cc)
}

/// @notice Enumerate valid NFTs
/// @dev Throws if `_index` >= `totalSupply()`.
/// @param _index A counter less than `totalSupply()`
/// @return The token identifier for the `_index`th NFT,
///  (sort order not specified)
func (f *front) TokenByIndex(cc *types.ContractContext, _index *big.Int) hash.Hash256 {
	return tokenByIndex(cc, _index)
}

/// @notice Enumerate NFTs assigned to an owner
/// @dev Throws if `_index` >= `balanceOf(_owner)` or if
///  `_owner` is the zero address, representing invalid NFTs.
/// @param _owner An address where we are interested in NFTs owned by them
/// @param _index A counter less than `balanceOf(_owner)`
/// @return The token identifier for the `_index`th NFT assigned to `_owner`,
///   (sort order not specified)
func (f *front) TokenOfOwnerByIndex(cc *types.ContractContext, _owner common.Address, _index *big.Int) hash.Hash256 {
	return tokenOfOwnerByIndex(cc, _owner, _index)
}

//////////////////////////////////////////////////
// Public Writer only owner Functions
//////////////////////////////////////////////////

func (f *front) Mint(cc *types.ContractContext, count *big.Int) ([]hash.Hash256, error) {
	return f.cont.mint(cc, count)
}

func (f *front) MintBatch(cc *types.ContractContext, addrs []common.Address) ([]hash.Hash256, error) {
	if cc.TargetHeight() < 13906097 {
		return nft721regacy.MintBatch(cc, f.cont.addr, addrs)
	}
	return f.cont.mintBatch(cc, addrs)
}

func (f *front) Burn(cc *types.ContractContext, nftID hash.Hash256) error {
	return burn(cc, nftID)
}

/// @notice A distinct Uniform Resource Identifier (URI) for a given asset.
/// @dev Throws if `_tokenId` is not a valid NFT. URIs are defined in RFC
///  3986. The URI may point to a JSON file that conforms to the "ERC721
///  Metadata JSON Schema".
func (f *front) SetBaseURI(cc *types.ContractContext, uri string) error {
	return f.cont.setBaseURI(cc, uri)
}

/// @notice A distinct Uniform Resource Identifier (URI) for a given asset.
/// @dev Throws if `_tokenId` is not a valid NFT. URIs are defined in RFC
///  3986. The URI may point to a JSON file that conforms to the "ERC721
///  Metadata JSON Schema".
func (f *front) SetTokenURI(cc *types.ContractContext, tokenID hash.Hash256, uri string) error {
	return f.cont.setTokenURI(cc, tokenID, uri)
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////

/// @notice Transfers the ownership of an NFT from one address to another address
/// @dev Throws unless `msg.sender` is the current owner, an authorized
///  operator, or the approved address for this NFT. Throws if `_from` is
///  not the current owner. Throws if `_to` is the zero address. Throws if
///  `_tokenId` is not a valid NFT. When transfer is complete, this function
///  checks if `_to` is a smart contract (code size > 0). If so, it calls
///  `onERC721Received` on `_to` and throws if the return value is not
///  `bytes4(keccak256("onERC721Received(address,address,uint256,bytes)"))`.
/// @param _from The current owner of the NFT
/// @param _to The new owner
/// @param _tokenId The NFT to transfer
/// @param data Additional data with no specified format, sent in call to `_to`
func (f *front) SafeTransferFrom(cc *types.ContractContext, _from common.Address, _to common.Address, _tokenId hash.Hash256, data []byte) error {
	return f.cont.safeTransferFrom(cc, _from, _to, _tokenId, data)
}

/// @notice Transfer ownership of an NFT -- THE CALLER IS RESPONSIBLE
///  TO CONFIRM THAT `_to` IS CAPABLE OF RECEIVING NFTS OR ELSE
///  THEY MAY BE PERMANENTLY LOST
/// @dev Throws unless `msg.sender` is the current owner, an authorized
///  operator, or the approved address for this NFT. Throws if `_from` is
///  not the current owner. Throws if `_to` is the zero address. Throws if
///  `_tokenId` is not a valid NFT.
/// @param _from The current owner of the NFT
/// @param _to The new owner
/// @param _tokenId The NFT to transfer
func (f *front) TransferFrom(cc *types.ContractContext, _from common.Address, _to common.Address, _tokenId hash.Hash256) error {
	return transferFrom(cc, _from, _to, _tokenId)
}

/// @notice Change or reaffirm the approved address for an NFT
/// @dev The zero address indicates there is no approved address.
///  Throws unless `msg.sender` is the current NFT owner, or an authorized
///  operator of the current owner.
/// @param _approved The new approved NFT controller
/// @param _tokenId The NFT to approve
func (f *front) Approve(cc *types.ContractContext, _approved common.Address, _tokenId hash.Hash256) error {
	return approve(cc, _approved, _tokenId)
}

/// @notice Enable or disable approval for a third party ("operator") to manage
///  all of `msg.sender`'s assets
/// @dev Emits the ApprovalForAll event. The contract MUST allow
///  multiple operators per owner.
/// @param _operator Address to add to the set of authorized operators
/// @param _approved True if the operator is approved, false to revoke approval
func (f *front) SetApprovalForAll(cc *types.ContractContext, _operator common.Address, _approved bool) {
	setApprovalForAll(cc, _operator, _approved)
}

func (f *front) PrintContractData(cc *types.ContractContext, addr common.Address) {
	printContractData(cc, addr)
}
