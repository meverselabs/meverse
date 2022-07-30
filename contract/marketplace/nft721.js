const tagOwner               = "01"
const tagMinter              = "02"
const tagNAME                = "03"
const tagSYMBOL              = "04"

const tagBaseURI             = "05"
const tagTokenURI            = "06"
const tagNFTCount            = "07"
const tagNFTMakeIndex        = "08"
const tagNFTIndex            = "09"
const tagIndexNFT            = "10"
const tagNFTOwner            = "11"
const tagTokenApprove        = "12"
const tagTokenApproveForAll  = "13"

//////////////////////////////////////////////////
// Public Functions
//////////////////////////////////////////////////

function Init(owner, _name, _symbol) {
    require(address(owner) !== address(0), "owner not provided")
	Mev.SetContractData(tagOwner, owner)
	Mev.SetContractData(tagMinter+BigInt(owner).toString(16), "1")
	Mev.SetContractData(tagNAME, _name)
	Mev.SetContractData(tagSYMBOL, _symbol)
}

/// @notice A descriptive name for a collection of NFTs in this contract
function name() {
	return Mev.ContractData(tagNAME)
}

/// @notice An abbreviated name for NFTs in this contract
function symbol() {
	return Mev.ContractData(tagSYMBOL)
}

function supportsInterface(interfaceID) {
	Mev.Log("supportsInterface", interfaceID)
    if (interfaceID == "[1 255 9 a7]") {
        return true
    }
    return false
}

/// @notice A distinct Uniform Resource Identifier (URI) for a given asset.
/// @dev Throws if `_tokenId` is not a valid NFT. URIs are defined in RFC
///  3986. The URI may point to a JSON file that conforms to the "ERC721
///  Metadata JSON Schema".
function tokenURI(_tokenId) {
	_tokenId = BigInt(_tokenId)
	let body = Mev.ContractData(_makeTokenURIKey(_tokenId))
	if (body == "") {
		body = baseURI()
	}
	_tokenId = _tokenId.toString(16)
    // if (_tokenId.indexOf("0x") == 0) {
    //     _tokenId = _tokenId.replace("0x", "")
    // }
    _tokenId = "0x"+_tokenId.padStart(64, '0')

	return body.replace("{id}", _tokenId)
}

function baseURI() {
	return Mev.ContractData(tagBaseURI)
}

/// @notice Count all NFTs assigned to an owner
/// @dev NFTs assigned to the zero address are considered invalid, and this
///  function throws for queries about the zero address.
/// @param _owner An address for whom to query the balance
/// @return The number of NFTs owned by `_owner`, possibly zero
function balanceOf(_owner)  {
	let bal = Mev.AccountData(_owner, tagNFTCount)
    if (bal == "") {
        bal = BigInt(0)
    } else {
        bal = BigInt(bal)
    }
	return bal
}

/// @notice Find the owner of an NFT
/// @dev NFTs assigned to zero address are considered invalid, and queries
///  about them do throw.
/// @param _tokenId The identifier for an NFT
/// @return The address of the owner of the NFT
function ownerOf(_tokenId) {
	_tokenId = BigInt(_tokenId)
	return address(Mev.ContractData(_makeNFTOwnerKey(_tokenId)))
}

/// @notice Get the approved address for a single NFT
/// @dev Throws if `_tokenId` is not a valid NFT.
/// @param _tokenId The NFT to find the approved address for
/// @return The approved address for this NFT, or the zero address if there is none
function getApproved(_tokenId) {
	_tokenId = BigInt(_tokenId)
    let addr = Mev.ContractData(_makeTokenApproveKey(_tokenId))
	return address(addr)
}

/// @notice Query if an address is an authorized operator for another address
/// @param _owner The address that owns the NFTs
/// @param _operator The address that acts on behalf of the owner
/// @return True if `_operator` is an approved operator for `_owner`, false otherwise
function isApprovedForAll(_owner, _operator) {
    let bs =  Mev.ContractData(_makeTokenApproveForAllKey(_owner, _operator))
	if (bs == "") {
		return false
	}
	return true
}

/// @notice Count NFTs tracked by this contract
/// @return A count of valid NFTs tracked by this contract, where each one of
///  them has an assigned and queryable owner not equal to the zero address
function totalSupply() {
	return "0x"+BigInt(Mev.ContractData(tagNFTCount)).toString(16)
}

/// @notice Enumerate valid NFTs
/// @dev Throws if `_index` >= `totalSupply()`.
/// @param _index A counter less than `totalSupply()`
/// @return The token identifier for the `_index`th NFT,
///  (sort order not specified)
function tokenByIndex(_index) {
    let inx = Mev.ContractData(_makeIndexNFTKey(_index))
	if (inx == "") {
		throw "not exist"
	}
	inx = BigInt(inx).toString(16)
	if (inx.length % 2 == 1) {
		return "0x0"+inx
	}
	return "0x"+inx
}

function tokenByRange(start, end) {
	start = BigInt(start)
	end = BigInt(end)

	if (start > end) {
		[start, end] = [end, start]
	}

	let total = BigInt(Mev.ContractData(tagNFTCount))-1n
	if (end > total) {
		end = total
	}
	if (end > total) {
		end = total
	}

	let list = []
	for (var i = start; i <= end ; i++) {
		let inx = Mev.ContractData(_makeIndexNFTKey(i))
		if (inx == "") {
			throw "not exist ("+i+") index nft"
		}
		inx = BigInt(inx).toString(16)
		if (inx.length % 2 == 1) {
			list.push("0x0"+inx)
		} else {
			list.push("0x"+inx)
		}
	}

	return list
}

/// @notice Enumerate NFTs assigned to an owner
/// @dev Throws if `_index` >= `balanceOf(_owner)` or if
///  `_owner` is the zero address, representing invalid NFTs.
/// @param _owner An address where we are interested in NFTs owned by them
/// @param _index A counter less than `balanceOf(_owner)`
/// @return The token identifier for the `_index`th NFT assigned to `_owner`,
///   (sort order not specified)
function tokenOfOwnerByIndex(_owner, _index) {
	let inx = Mev.AccountData(address(_owner), _makeIndexNFTKey(_index))
	if (inx == "") {
		throw "not exist"
	}
	inx = BigInt(inx).toString(16)
	if (inx.length % 2 == 1) {
		return "0x0"+inx
	}
	return "0x"+inx
}

function tokenOfOwnerByRange(_owner, start, end) {
	start = BigInt(start)
	end = BigInt(end)

	if (start > end) {
		[start, end] = [end, start]
	}

	
	let total = balanceOf(_owner)-1n
	if (end > total) {
		end = total
	}
	if (end > total) {
		end = total
	}

	let list = []
	for (var i = start; i <= end ; i++) {

		let inx = Mev.AccountData(address(_owner), _makeIndexNFTKey(i))
		if (inx == "") {
			throw "not exist ("+i+") index nft"
		}
		inx = BigInt(inx).toString(16)
		if (inx.length % 2 == 1) {
			list.push("0x0"+inx)
		} else {
			list.push("0x"+inx)
		}
	}
	return list
}

//////////////////////////////////////////////////
// Public Write only owner Functions
//////////////////////////////////////////////////
function isOwner() {
	let addr = Mev.ContractData(tagOwner)
	return address(Mev.From()) == address(addr)
}
function isMinter() {
	let isMinter = Mev.ContractData(tagMinter+BigInt(Mev.From()).toString(16))
	return isMinter == "1"
}
function setMinter(addr) {
    require(isOwner(), "doesn't have mint permission")
	Mev.SetContractData(tagMinter+BigInt(addr).toString(16), "1")
}
function revertMinter(addr) {
    require(isOwner(), "doesn't have mint permission")
	Mev.SetContractData(tagMinter+BigInt(addr).toString(16), "")
}

function mint(count) {
    require(isMinter(), "doesn't have mint permission")
    count = BigInt(count)
    require(count > 0, "mint count must over then 0")

	let nc = BigInt(Mev.ContractData(tagNFTCount))
	let limit = count + nc
	let hs = []
	while (nc < limit) {
		let nftID = _mintNFTWithIndex(Mev.This(), nc)
		_addNFTAccount(Mev.From(), nftID)
		nc++
		hs.push(nftID)
	}
	return JSON.stringify(hs)
}

function addr() {
	return Mev.This()
}

function getNftId(nc) {
	let seedAddr = Mev.This()
	let key = BigInt(nc)
    var hex = key.toString(16);
    if (hex.length % 2) {
      hex = '0' + hex;
    }
	let o = seedAddr+hex
	let o2 = Mev.Hash256(o)
	let o3 = Mev.Hash256(o2)
	let nftID = Mev.DoubleHash256(o)
	return seedAddr+":"+hex+":"+o+":"+nc+":"+key+":"+nftID+":"+o2+":"+o3+":"+BigInt(nftID)
}

function mintWithID(nftID) {
    require(isMinter(), "doesn't have mint permission")

	let nc = BigInt(Mev.ContractData(tagNFTCount))
	nftID = _mintNFTWithIndex(Mev.This(), nc, nftID)
	_addNFTAccount(Mev.From(), nftID)
	return nftID
}

function mintBatch(addrsStg) {
    let addrs = JSON.parse(addrsStg)
    require(isMinter(), "doesn't have mint permission")
    count = BigInt(addrs.length)
    require(count > 0, "mint count must over then 0")
    
	let hs = mint(count)
	hs = JSON.parse(hs)
    require(hs.length == addrs.length, "not enough mint count")
	for (var i = 0 ; i < addrs.length ; i++) {
		transferFrom(Mev.From(), addrs[i], hs[i])
	}
	return JSON.stringify(hs)
}

function _makeNFTID(seedAddr, nftID) {
	let key = BigInt(Mev.ContractData(tagNFTMakeIndex))
	key++

    var hex = key.toString(16);
    if (hex.length % 2) {
      hex = '0' + hex;
    }
	if (typeof nftID === "undefined") {
		nftID = Mev.DoubleHash256(seedAddr+hex)
	}

    Mev.SetContractData(tagNFTMakeIndex, key)
	return BigInt(nftID)
}

function _mintNFTWithIndex(seedAddr, nc, nftID) {
    nftID = _makeNFTID(seedAddr, nftID)

	let indexbs = Mev.ContractData(_makeIndexNFTKey(nc))
    require(indexbs == "", "try mint duplicate index")
    let mintKey = _makeNFTIndexKey(nftID)
	let nftbs = Mev.ContractData(mintKey)
    require(nftbs == "", "try mint duplicate nft")

	_setNFTIndex(nftID, nc)
	Mev.SetContractData(tagNFTCount, nc + 1n)
	return nftID
}

function _setNFTIndex(nftID, nc) {
    nc = BigInt(nc)
	Mev.SetContractData(_makeNFTIndexKey(nftID), nc)
	Mev.SetContractData(_makeIndexNFTKey(nc), nftID)
}

function _addNFTAccount(addr, nftID) {
    addr = address(addr)
	let nc = BigInt(Mev.AccountData(addr, tagNFTCount))

	let indexbs = Mev.AccountData(addr, _makeIndexNFTKey(nc))
    require(indexbs == "", "try mint duplicate index")
	let nftbs = Mev.AccountData(addr, _makeNFTIndexKey(nftID))
    require(nftbs == "", "try mint duplicate nft")

	_setNFTAccountIndex(addr, nftID, nc)
	Mev.SetAccountData(addr, tagNFTCount, nc + 1n)

	Mev.SetContractData(_makeNFTOwnerKey(nftID), addr)
}

function _setNFTAccountIndex(addr, nftID, nc) {
    nc = BigInt(nc)
	Mev.SetAccountData(addr, _makeNFTIndexKey(nftID), nc)
	Mev.SetAccountData(addr, _makeIndexNFTKey(nc), nftID)
}

function burn(nftID) {
    require(isMinter(), "doesn't have burn permission")
    
    let burnkey = _makeNFTIndexKey(nftID)
    let indexbs = Mev.ContractData(burnkey)
    require(indexbs != "", "not exist nft")

    let burnIndex = BigInt(indexbs)

	_removeNFT(nftID, burnIndex)
	return _removeNFTAccount(nftID)
}

function _removeNFT(nftID, burnIndex ) {
    burnIndex = BigInt(burnIndex)
	let lastIndex = totalSupply()
    lastIndex = BigInt(lastIndex)-1n

	Mev.SetContractData(_makeNFTIndexKey(nftID), "")

	if (burnIndex != lastIndex) {
		let lastNFTID = BigInt(tokenByIndex(lastIndex))
		_setNFTIndex(lastNFTID, burnIndex)
	}
	Mev.SetContractData(_makeIndexNFTKey(lastIndex), "")
	Mev.SetContractData(tagNFTCount, lastIndex)
}

function _removeNFTAccount(nftID) {
	let addr = ownerOf(nftID)

	let indexbs = Mev.AccountData(addr, _makeNFTIndexKey(nftID))
    require(indexbs != "", "not exist nft")
	let burnIndex = BigInt(indexbs)

	let lastIndex = balanceOf(addr)
    lastIndex = lastIndex - 1n

	Mev.SetAccountData(addr, _makeNFTIndexKey(nftID), "")

    if (burnIndex != lastIndex) {
		let lastNFTID = BigInt(tokenOfOwnerByIndex(addr, lastIndex))
		_setNFTAccountIndex(addr, lastNFTID, burnIndex)
	}
	Mev.SetContractData(_makeNFTOwnerKey(nftID), address(0))
	Mev.SetAccountData(addr, _makeIndexNFTKey(lastIndex), "")
	Mev.SetAccountData(addr, tagNFTCount, lastIndex)
}

function setBaseURI(uri) {
    require(isOwner(), "doesn't have setBaseURI permission")
	Mev.SetContractData(tagBaseURI, uri)
}

function setTokenURI(tokenID, uri) {
    require(isOwner(), "doesn't have setTokenURI permission")
	Mev.SetContractData(_makeTokenURIKey(tokenID), uri)
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
function safeTransferFrom(_from, _to, _tokenId, data) {
    transferFrom(_from, _to, _tokenId)
    let res = Mev.Exec(_to, "OnERC721Received", [Mev.This(), _from, _tokenId, data])
    require(res == "150b7a02", "onERC721Received invalid return" + res)
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
function transferFrom(_from, _to, _tokenId) {
	_tokenId = BigInt(_tokenId)
	_from = address(_from)
	_to = address(_to)
    _checkTransferPermission(_tokenId, _from, _to)
    _removeNFTAccount(_tokenId)
    _addNFTAccount(_to, _tokenId)
}

function _checkTransferPermission(_tokenId, _from, _to) {
    require(address(_to) != address(0), "to is the zero address")
    let owner = ownerOf(_tokenId)
    require(address(_from) == owner, "from is not the current owner")
    
    let from = address(Mev.From())
	if (from == owner) { // Throws unless `msg.sender` is the current owner
		return
	}
	if (getApproved(_tokenId) == from) { // or the approved address for this NFT.
		return
	}
	if (isApprovedForAll(owner, from)) { // an authorized operator
		return
	}
	throw "no permission"
}

/// @notice Change or reaffirm the approved address for an NFT
/// @dev The zero address indicates there is no approved address.
///  Throws unless `msg.sender` is the current NFT owner, or an authorized
///  operator of the current owner.
/// @param _approved The new approved NFT controller
/// @param _tokenId The NFT to approve
function approve(_approved, _tokenId) {
    require(ownerOf(_tokenId) == address(Mev.From()), "not token owner")
	Mev.SetContractData(_makeTokenApproveKey(_tokenId), address(_approved))
}

/// @notice Enable or disable approval for a third party ("operator") to manage
///  all of `msg.sender`'s assets
/// @dev Emits the ApprovalForAll event. The contract MUST allow
///  multiple operators per owner.
/// @param _operator Address to add to the set of authorized operators
/// @param _approved True if the operator is approved, false to revoke approval
function setApprovalForAll(_operator, _approved) {
	if (_approved === true || _approved == "true" ) {
        let key = _makeTokenApproveForAllKey(Mev.From(), _operator)
		Mev.SetContractData(key, 1)
	} else {
		Mev.SetContractData(_makeTokenApproveForAllKey(Mev.From(), _operator), "")
	}
}

function PrintContractData(addr) {
    let ts = totalSupply()
	Mev.Log("printContractData start")
	for (var i = BigInt(0); i < ts; i++) {
		Mev.Log(tokenByIndex(i))
	}
	Mev.Log("printContractData end and addr start")
    let bo = balanceOf(addr)
	for (var i = BigInt(0); i < bo; i++) {
		Mev.Log(tokenOfOwnerByIndex(addr, i))
	}
	Mev.Log("printContractData end")
}


/********************************
 * 
 * keys
 * 
 *********************************/

function _makeNFTKey(key, body) {
	return key+""+BigInt(body).toString(16)
}

function _makeNFTIndexKey(k) {
    k = BigInt(k)
	return _makeNFTKey(tagNFTIndex, k)
}
function _makeIndexNFTKey(i) {
	return _makeNFTKey(tagIndexNFT, BigInt(i))
}
function _makeNFTOwnerKey(tokenID) {
	return _makeNFTKey(tagNFTOwner, BigInt(tokenID))
}
function _makeTokenApproveKey(i) {
    i = BigInt(i)
	return _makeNFTKey(tagTokenApprove, i)
}
function _makeTokenApproveForAllKey(_owner, _operator) {
	return _makeNFTKey(tagTokenApproveForAll, BigInt(_owner)*10000000000000000000000000000000000000000n + BigInt(_operator))
}
function _makeTokenURIKey(tokenID) {
	tokenID = BigInt(tokenID)
	return _makeNFTKey(tagTokenURI, tokenID)
}
