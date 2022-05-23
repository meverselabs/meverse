package nft721

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
)

type NFT721Contract struct {
	addr   common.Address
	master common.Address
}

// func (cont *NFT721Contract) Name() string {
// 	return "NFT721"
// }

func (cont *NFT721Contract) Address() common.Address {
	return cont.addr
}

func (cont *NFT721Contract) Master() common.Address {
	return cont.master
}

func (cont *NFT721Contract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

func (cont *NFT721Contract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &NFT721ContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}

	cc.SetContractData([]byte{tagOwner}, data.Owner[:])
	cc.SetContractData([]byte{tagName}, []byte(data.Name))
	cc.SetContractData([]byte{tagSymbol}, []byte(data.Symbol))
	return nil
}

func (cont *NFT721Contract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

//////////////////////////////////////////////////
// Public Read Functions
//////////////////////////////////////////////////
// bytes4 constant public ERC1155_ERC165 = 0xd9b67a26; // ERC-165 identifier for the main token standard.
// bytes4 constant public ERC1155_ERC165_TOKENRECEIVER = 0x4e2312e0; // ERC-165 identifier for the `ERC1155TokenReceiver` support (i.e. `bytes4(keccak256("onERC1155Received(address,address,uint256,uint256,bytes)")) ^ bytes4(keccak256("onERC1155BatchReceived(address,address,uint256[],uint256[],bytes)"))`).
// bytes4 constant public ERC1155_ACCEPTED = 0xf23a6e61; // Return value from `onERC1155Received` call if a contract accepts receipt (i.e `bytes4(keccak256("onERC1155Received(address,address,uint256,uint256,bytes)"))`).
// bytes4 constant public ERC1155_BATCH_ACCEPTED = 0xbc197c81; // Return value from `onERC1155BatchReceived` call if a contract accepts receipt (i.e `bytes4(keccak256("onERC1155BatchReceived(address,address,uint256[],uint256[],bytes)"))`).

func SupportsInterface(cc *types.ContractContext, interfaceID []byte) bool {
	switch 0 {
	case bytes.Compare(interfaceID, []byte{0x01, 0xff, 0xc9, 0xa7}):
		return true
	}
	return false
}

//////////////////////////////////////////////////
// Public Read Functions
//////////////////////////////////////////////////

/// @notice A descriptive name for a collection of NFTs in this contract
func name(cc *types.ContractContext) string {
	bs := cc.ContractData([]byte{tagName})
	return string(bs)
}

/// @notice An abbreviated name for NFTs in this contract
func symbol(cc *types.ContractContext) string {
	bs := cc.ContractData([]byte{tagSymbol})
	return string(bs)
}

/// @notice A distinct Uniform Resource Identifier (URI) for a given asset.
/// @dev Throws if `_tokenId` is not a valid NFT. URIs are defined in RFC
///  3986. The URI may point to a JSON file that conforms to the "ERC721
///  Metadata JSON Schema".
func (cont *NFT721Contract) tokenURI(cc *types.ContractContext, _tokenId hash.Hash256) string {
	strID := hex.EncodeToString(_tokenId.Bytes())
	idstr := fmt.Sprintf("%064v", strID)
	body := ""

	bs := cc.ContractData(makeTokenURIKey(_tokenId))
	if len(bs) != 0 {
		body = string(bs)
	} else {
		body = cont.baseURI(cc)
	}
	return strings.Replace(body, "{id}", idstr, -1)
}

func (cont *NFT721Contract) baseURI(cc *types.ContractContext) string {
	bs := cc.ContractData([]byte{tagBaseURI})
	if len(bs) == 0 {
		return ""
	}
	return string(bs)
}

/// @notice Count all NFTs assigned to an owner
/// @dev NFTs assigned to the zero address are considered invalid, and this
///  function throws for queries about the zero address.
/// @param _owner An address for whom to query the balance
/// @return The number of NFTs owned by `_owner`, possibly zero
func balanceOf(cc *types.ContractContext, _owner common.Address) *big.Int {
	bs := cc.AccountData(_owner, []byte{tagNFTCount})
	return big.NewInt(0).SetBytes(bs)
}

/// @notice Find the owner of an NFT
/// @dev NFTs assigned to zero address are considered invalid, and queries
///  about them do throw.
/// @param _tokenId The identifier for an NFT
/// @return The address of the owner of the NFT
func ownerOf(cc *types.ContractContext, _tokenId hash.Hash256) common.Address {
	bs := cc.ContractData(makeNFTOwnerKey(_tokenId))
	return common.BytesToAddress(bs)
}

/// @notice Get the approved address for a single NFT
/// @dev Throws if `_tokenId` is not a valid NFT.
/// @param _tokenId The NFT to find the approved address for
/// @return The approved address for this NFT, or the zero address if there is none
func getApproved(cc *types.ContractContext, _tokenId hash.Hash256) common.Address {
	bs := cc.ContractData(makeTokenApproveKey(_tokenId))
	return common.BytesToAddress(bs)
}

/// @notice Query if an address is an authorized operator for another address
/// @param _owner The address that owns the NFTs
/// @param _operator The address that acts on behalf of the owner
/// @return True if `_operator` is an approved operator for `_owner`, false otherwise
func isApprovedForAll(cc *types.ContractContext, _owner common.Address, _operator common.Address) bool {
	bs := cc.ContractData(makeTokenApproveForAllKey(_owner, _operator))
	if len(bs) == 0 {
		return false
	}
	return true
}

/// @notice Count NFTs tracked by this contract
/// @return A count of valid NFTs tracked by this contract, where each one of
///  them has an assigned and queryable owner not equal to the zero address
func totalSupply(cc *types.ContractContext) *big.Int {
	bs := cc.ContractData([]byte{tagNFTCount})
	return big.NewInt(0).SetBytes(bs)
}

/// @notice Enumerate valid NFTs
/// @dev Throws if `_index` >= `totalSupply()`.
/// @param _index A counter less than `totalSupply()`
/// @return The token identifier for the `_index`th NFT,
///  (sort order not specified)
func tokenByIndex(cc *types.ContractContext, _index *big.Int) hash.Hash256 {
	lastNFTbs := cc.ContractData(makeIndexNFTKey(_index))
	lastNFTID := hash.Hash256{}
	copy(lastNFTID[:], lastNFTbs)
	return lastNFTID
}

/// @notice Enumerate NFTs assigned to an owner
/// @dev Throws if `_index` >= `balanceOf(_owner)` or if
///  `_owner` is the zero address, representing invalid NFTs.
/// @param _owner An address where we are interested in NFTs owned by them
/// @param _index A counter less than `balanceOf(_owner)`
/// @return The token identifier for the `_index`th NFT assigned to `_owner`,
///   (sort order not specified)
func tokenOfOwnerByIndex(cc *types.ContractContext, _owner common.Address, _index *big.Int) hash.Hash256 {
	lastNFTbs := cc.AccountData(_owner, makeIndexNFTKey(_index))
	lastNFTID := hash.Hash256{}
	copy(lastNFTID[:], lastNFTbs)
	return lastNFTID
}

//////////////////////////////////////////////////
// Public Write only owner Functions
//////////////////////////////////////////////////
func isOwner(cc *types.ContractContext) bool {
	bs := cc.ContractData([]byte{tagOwner})
	return cc.From() == common.BytesToAddress(bs)
}

func (cont *NFT721Contract) mint(cc *types.ContractContext, count *big.Int) ([]hash.Hash256, error) {
	if !isOwner(cc) {
		return nil, errors.New("doesn't have mint permission")
	}

	if len(count.Bytes()) == 0 {
		return nil, errors.New("mint count must over then 0")
	}

	bs := cc.ContractData([]byte{tagNFTCount})
	nc := big.NewInt(0).SetBytes(bs)

	limit := big.NewInt(0).Add(nc, count)

	one := big.NewInt(1)
	hs := []hash.Hash256{}
	for nc.Cmp(limit) < 0 {
		nftID, err := _mintNFTWithIndex(cc, cont.addr, nc)
		if err != nil {
			return []hash.Hash256{}, err
		}
		err = addNFTAccount(cc, cc.From(), nftID)
		if err != nil {
			return []hash.Hash256{}, err
		}

		nc.Add(nc, one)
		hs = append(hs, nftID)
	}

	return hs, nil
}

func (cont *NFT721Contract) mintBatch(cc *types.ContractContext, addrs []common.Address) ([]hash.Hash256, error) {
	if !isOwner(cc) {
		return nil, errors.New("doesn't have mint permission")
	}

	if len(addrs) == 0 {
		return nil, errors.New("mint count must over then 0")
	}

	count := big.NewInt(int64(len(addrs)))
	hs, err := cont.mint(cc, count)
	if err != nil {
		return nil, err
	}
	for i, addr := range addrs {
		transferFrom(cc, cc.From(), addr, hs[i])
	}
	return hs, nil
}

func makeNFTID(cc *types.ContractContext, seedAddr common.Address) hash.Hash256 {
	keyBs := cc.ContractData([]byte{tagNFTMakeIndex})
	key := big.NewInt(0).SetBytes(keyBs)
	key.Add(key, big.NewInt(1))

	bs := append(seedAddr[:], key.Bytes()...)
	nftID := hash.DoubleHash(bs)

	cc.SetContractData([]byte{tagNFTMakeIndex}, key.Bytes())
	return nftID
}

func _mintNFTWithIndex(cc *types.ContractContext, seedAddr common.Address, nc *big.Int) (hash.Hash256, error) {
	bs := append(seedAddr[:], nc.Bytes()...)
	nftID := hash.DoubleHash(bs)

	indexbs := cc.ContractData(makeIndexNFTKey(nc))
	if len(indexbs) != 0 {
		return hash.Hash256{}, errors.New("try mint duplicate index")
	}
	nftbs := cc.ContractData(makeNFTIndexKey(nftID))
	if len(nftbs) != 0 {
		return hash.Hash256{}, errors.New("try mint duplicate nft")
	}

	setNFTIndex(cc, nftID, nc)
	cc.SetContractData([]byte{tagNFTCount}, big.NewInt(0).Add(nc, big.NewInt(1)).Bytes())
	return nftID, nil
}

func setNFTIndex(cc *types.ContractContext, nftID hash.Hash256, nc *big.Int) {
	nbs := nc.Bytes()
	if len(nbs) == 0 {
		nbs = []byte{0}
	}
	cc.SetContractData(makeNFTIndexKey(nftID), nbs)
	cc.SetContractData(makeIndexNFTKey(nc), nftID.Bytes())
}

func addNFTAccount(cc *types.ContractContext, addr common.Address, nftID hash.Hash256) error {
	bs := cc.AccountData(addr, []byte{tagNFTCount})
	nc := big.NewInt(0).SetBytes(bs)

	indexbs := cc.AccountData(addr, makeIndexNFTKey(nc))
	if len(indexbs) != 0 {
		return errors.New("try mint duplicate index")
	}
	nftbs := cc.AccountData(addr, makeNFTIndexKey(nftID))
	if len(nftbs) != 0 {
		return errors.New("try mint duplicate nft")
	}

	setNFTAccountIndex(cc, addr, nftID, nc)
	cc.SetAccountData(addr, []byte{tagNFTCount}, big.NewInt(0).Add(nc, big.NewInt(1)).Bytes())

	cc.SetContractData(makeNFTOwnerKey(nftID), addr[:])
	return nil
}

func setNFTAccountIndex(cc *types.ContractContext, addr common.Address, nftID hash.Hash256, nc *big.Int) {
	nbs := nc.Bytes()
	if len(nbs) == 0 {
		nbs = []byte{0}
	}
	cc.SetAccountData(addr, makeNFTIndexKey(nftID), nbs)
	cc.SetAccountData(addr, makeIndexNFTKey(nc), nftID.Bytes())
}

func burn(cc *types.ContractContext, nftID hash.Hash256) error {
	if !isOwner(cc) {
		return errors.New("doesn't have mint permission")
	}

	indexbs := cc.ContractData(makeNFTIndexKey(nftID))
	if len(indexbs) == 0 {
		return errors.New("not exist nft")
	}
	burnIndex := big.NewInt(0).SetBytes(indexbs)

	removeNFT(cc, nftID, burnIndex)
	return removeNFTAccount(cc, nftID)
}

func removeNFT(cc *types.ContractContext, nftID hash.Hash256, burnIndex *big.Int) {
	lastIndex := totalSupply(cc)
	lastIndex.Sub(lastIndex, big.NewInt(1))

	cc.SetContractData(makeNFTIndexKey(nftID), nil)

	if burnIndex.Cmp(lastIndex) != 0 {
		lastNFTID := tokenByIndex(cc, lastIndex)
		setNFTIndex(cc, lastNFTID, burnIndex)
	}
	cc.SetContractData(makeIndexNFTKey(lastIndex), nil)
	cc.SetContractData([]byte{tagNFTCount}, lastIndex.Bytes())
}

func removeNFTAccount(cc *types.ContractContext, nftID hash.Hash256) error {
	addr := ownerOf(cc, nftID)

	indexbs := cc.AccountData(addr, makeNFTIndexKey(nftID))
	if len(indexbs) == 0 {
		return errors.New("not exist nft")
	}
	burnIndex := big.NewInt(0).SetBytes(indexbs)

	lastIndex := balanceOf(cc, addr)
	lastIndex.Sub(lastIndex, big.NewInt(1))

	cc.SetAccountData(addr, makeNFTIndexKey(nftID), nil)

	if burnIndex.Cmp(lastIndex) != 0 {
		lastNFTID := tokenOfOwnerByIndex(cc, addr, lastIndex)
		setNFTAccountIndex(cc, addr, lastNFTID, burnIndex)
	}
	cc.SetAccountData(addr, makeIndexNFTKey(lastIndex), nil)
	cc.SetAccountData(addr, []byte{tagNFTCount}, lastIndex.Bytes())
	return nil
}

func (cont *NFT721Contract) setBaseURI(cc *types.ContractContext, uri string) error {
	if !isOwner(cc) {
		return errors.New("doesn't have setBaseURI permission")
	}

	cc.SetContractData([]byte{tagBaseURI}, []byte(uri))
	return nil
}

func (cont *NFT721Contract) setTokenURI(cc *types.ContractContext, tokenID hash.Hash256, uri string) error {
	if !isOwner(cc) {
		return errors.New("doesn't have setTokenURI permission")
	}

	cc.SetContractData(makeTokenURIKey(tokenID), []byte(uri))
	return nil
}

//////////////////////////////////////////////////
// Public Write Functions
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
func (cont *NFT721Contract) safeTransferFrom(cc *types.ContractContext, _from common.Address, _to common.Address, _tokenId hash.Hash256, data []byte) error {
	if err := transferFrom(cc, _from, _to, _tokenId); err != nil {
		return err
	}

	if in, err := cc.Exec(cc, _to, "OnERC721Received", []interface{}{cont.addr, _from, _tokenId, data}); err != nil {
		return err
	} else if len(in) == 0 {
		return errors.New("OnERC721Received invalid return")
	} else if bs, ok := in[0].([]byte); ok && bytes.Compare(bs, []byte{0x15, 0x0b, 0x7a, 0x02}) == 0 {

	}

	return nil
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
func transferFrom(cc *types.ContractContext, _from common.Address, _to common.Address, _tokenId hash.Hash256) error {
	if err := checkTransferPermission(cc, _tokenId, _from, _to); err != nil {
		return err
	}

	if err := removeNFTAccount(cc, _tokenId); err != nil {
		return err
	}

	if err := addNFTAccount(cc, _to, _tokenId); err != nil {
		return err
	}

	return nil
}

func checkTransferPermission(cc *types.ContractContext, _tokenId hash.Hash256, _from, _to common.Address) error {
	if _to == common.ZeroAddr { // Throws if `_to` is the zero address.
		return errors.New("to is the zero address")
	}

	owner := ownerOf(cc, _tokenId)
	if _from != owner { // Throws if `_from` is not the current owner.
		return errors.New("from is not the current owner")
	}

	from := cc.From()
	if owner == from { // Throws unless `msg.sender` is the current owner
		return nil
	}

	if getApproved(cc, _tokenId) == from { // or the approved address for this NFT.
		return nil
	}
	if isApprovedForAll(cc, owner, cc.From()) { // an authorized operator
		return nil
	}
	return errors.New("no permission")
}

/// @notice Change or reaffirm the approved address for an NFT
/// @dev The zero address indicates there is no approved address.
///  Throws unless `msg.sender` is the current NFT owner, or an authorized
///  operator of the current owner.
/// @param _approved The new approved NFT controller
/// @param _tokenId The NFT to approve
func approve(cc *types.ContractContext, _approved common.Address, _tokenId hash.Hash256) error {
	if ownerOf(cc, _tokenId) != cc.From() {
		return errors.New("not token owner")
	}
	cc.SetContractData(makeTokenApproveKey(_tokenId), _approved[:])
	return nil
}

/// @notice Enable or disable approval for a third party ("operator") to manage
///  all of `msg.sender`'s assets
/// @dev Emits the ApprovalForAll event. The contract MUST allow
///  multiple operators per owner.
/// @param _operator Address to add to the set of authorized operators
/// @param _approved True if the operator is approved, false to revoke approval
func setApprovalForAll(cc *types.ContractContext, _operator common.Address, _approved bool) {
	if _approved {
		cc.SetContractData(makeTokenApproveForAllKey(cc.From(), _operator), []byte{1})
	} else {
		cc.SetContractData(makeTokenApproveForAllKey(cc.From(), _operator), nil)
	}
}

func printContractData(cc *types.ContractContext, addr common.Address) {
	log.Println("printContractData start")
	for i := int64(0); i < 10; i++ {
		bi := big.NewInt(i)
		hs := tokenByIndex(cc, bi)
		log.Println(hs.String())
	}
	log.Println("printContractData end and addr start")
	for i := int64(0); i < 10; i++ {
		bi := big.NewInt(i)
		hs := tokenOfOwnerByIndex(cc, addr, bi)
		log.Println(hs.String())
	}
	log.Println("printContractData end")
}
