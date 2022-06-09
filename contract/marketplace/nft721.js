const tagOwner               = "01"
const tagNAME                = "02"
const tagSYMBOL              = "03"

const tagBaseURI             = "04"
const tagTokenURI            = "05"
const tagNFTCount            = "06"
const tagNFTMakeIndex        = "07"
const tagNFTIndex            = "08"
const tagIndexNFT            = "09"
const tagNFTOwner            = "10"
const tagTokenApprove        = "11"
const tagTokenApproveForAll  = "12"

// function require(condition, reson) {
//     if (!condition) {
//         throw reson
//     }
// }

// function address(addrRow) {
//     let bigIntAddr;
//     switch (typeof addrRow) {
//         case "string":
//             if (addrRow.indexOf("0x") == -1 && addrRow != "") {
//                 addrRow = "0x" + addrRow
//             }
//             bigIntAddr = BigInt(addrRow);
//             break;
//         case "bigint":
//             bigIntAddr = addrRow
//             break;
//         case "number":
//             bigIntAddr = BigInt(addrRow)
//             break;
//         default:
//             throw `invalid address type(${typeof addrRow}, ${addrRow})`
//     }

//     if (bigIntAddr < 0) {
//         throw "invalid address value" + addrRow
//     }
//     var hex = bigIntAddr.toString(16);
//     if (hex.length % 2) {
//         hex = '0' + hex;
//     }

//     return "0x"+hex.padStart(40, '0');
// }

// function _msgSender() {
//     return Mev.From()
// }


//////////////////////////////////////////////////
// Public Functions
//////////////////////////////////////////////////

function Init(owner, _name, _symbol) {
    require(address(owner) !== address(0), "owner not provided")
	Mev.SetContractData(tagOwner, owner)
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

/// @notice An abbreviated name for NFTs in this contract
function tesrsrer() {
	return "symboldsfawef"
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
	_tokenId = Bigint(_tokenId)
	let body = Mev.ContractData(makeTokenURIKey(_tokenId))
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
	return address(Mev.ContractData(makeNFTOwnerKey(_tokenId)))
}

/// @notice Get the approved address for a single NFT
/// @dev Throws if `_tokenId` is not a valid NFT.
/// @param _tokenId The NFT to find the approved address for
/// @return The approved address for this NFT, or the zero address if there is none
function getApproved(_tokenId) {
	_tokenId = BigInt(_tokenId)
    let addr = Mev.ContractData(makeTokenApproveKey(_tokenId))
	return address(addr)
}

/// @notice Query if an address is an authorized operator for another address
/// @param _owner The address that owns the NFTs
/// @param _operator The address that acts on behalf of the owner
/// @return True if `_operator` is an approved operator for `_owner`, false otherwise
function isApprovedForAll(_owner, _operator) {
    let bs =  Mev.ContractData(makeTokenApproveForAllKey(_owner, _operator))
	if (bs == "") {
		return false
	}
	return true
}

/// @notice Count NFTs tracked by this contract
/// @return A count of valid NFTs tracked by this contract, where each one of
///  them has an assigned and queryable owner not equal to the zero address
function totalSupply()  {
	return BigInt(Mev.ContractData(tagNFTCount))
}

/// @notice Enumerate valid NFTs
/// @dev Throws if `_index` >= `totalSupply()`.
/// @param _index A counter less than `totalSupply()`
/// @return The token identifier for the `_index`th NFT,
///  (sort order not specified)
function tokenByIndex(_index) {
    let inx = Mev.ContractData(makeIndexNFTKey(_index))
	return BigInt(inx)
}

/// @notice Enumerate NFTs assigned to an owner
/// @dev Throws if `_index` >= `balanceOf(_owner)` or if
///  `_owner` is the zero address, representing invalid NFTs.
/// @param _owner An address where we are interested in NFTs owned by them
/// @param _index A counter less than `balanceOf(_owner)`
/// @return The token identifier for the `_index`th NFT assigned to `_owner`,
///   (sort order not specified)
function tokenOfOwnerByIndex(_owner, _index ) {
	return BigInt(Mev.AccountData(_owner, makeIndexNFTKey(_index)))
}

//////////////////////////////////////////////////
// Public Write only owner Functions
//////////////////////////////////////////////////
function isOwner() {
	let addr = Mev.ContractData(tagOwner)
	return address(Mev.From()) == address(addr)
}

function mint(count) {
    require(isOwner(), "doesn't have mint permission")
    count = BigInt(count)
    require(count > 0, "mint count must over then 0")

	let nc = BigInt(Mev.ContractData(tagNFTCount))
	let limit = count + nc
	let hs = []
	while (nc < limit) {
		let nftID = _mintNFTWithIndex(Mev.This(), nc)
		addNFTAccount(Mev.From(), nftID)
		nc++
		hs.push(nftID)
	}
	return hs
}

function mintWithID(nftID) {
    require(isOwner(), "doesn't have mint permission")

	let nc = BigInt(Mev.ContractData(tagNFTCount))
	let hs = []
	nftID = _mintNFTWithIndex(Mev.This(), nc, nftID)
	addNFTAccount(Mev.From(), nftID)
	hs.push(nftID)
	return hs
}

function mintBatch(addrsStg) {
    let addrs = JSON.parse(addrsStg)
    require(isOwner(), "doesn't have mint permission")
    count = BigInt(addrs.length)
    require(count > 0, "mint count must over then 0")
    
	let hs = mint(count)
    require(hs.length == addrs.length, "not enough mint count")
	for (var i = 0 ; i < addrs.length ; i++) {
		transferFrom(Mev.From(), addrs[i], hs[i])
	}
	return hs
}

function makeNFTID(seedAddr, nftID) {
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
    nftID = makeNFTID(seedAddr, nftID)

	let indexbs = Mev.ContractData(makeIndexNFTKey(nc))
    require(indexbs == "", "try mint duplicate index")
    let mintKey = makeNFTIndexKey(nftID)
	let nftbs = Mev.ContractData(mintKey)
    require(nftbs == "", "try mint duplicate nft")

	setNFTIndex(nftID, nc)
	Mev.SetContractData(tagNFTCount, nc + 1n)
	return nftID
}

function setNFTIndex(nftID, nc) {
    nc = BigInt(nc)
	Mev.SetContractData(makeNFTIndexKey(nftID), nc)
	Mev.SetContractData(makeIndexNFTKey(nc), nftID)
}

function addNFTAccount(addr, nftID) {
    addr = address(addr)
	let nc = BigInt(Mev.AccountData(addr, tagNFTCount))

	let indexbs = Mev.AccountData(addr, makeIndexNFTKey(nc))
    require(indexbs == "", "try mint duplicate index")
	let nftbs = Mev.AccountData(addr, makeNFTIndexKey(nftID))
    require(nftbs == "", "try mint duplicate nft")

	setNFTAccountIndex(addr, nftID, nc)
	Mev.SetAccountData(addr, tagNFTCount, nc + 1n)

	Mev.SetContractData(makeNFTOwnerKey(nftID), addr)
}

function setNFTAccountIndex(addr, nftID, nc) {
    nc = BigInt(nc)
	Mev.SetAccountData(addr, makeNFTIndexKey(nftID), nc)
	Mev.SetAccountData(addr, makeIndexNFTKey(nc), nftID)
}

function burn(nftID) {
    require(isOwner(), "doesn't have burn permission")
    
    let burnkey = makeNFTIndexKey(nftID)
    let indexbs = Mev.ContractData(burnkey)
    require(indexbs != "", "not exist nft")

    let burnIndex = BigInt(indexbs)

	removeNFT(nftID, burnIndex)
	return removeNFTAccount(nftID)
}

function removeNFT(nftID, burnIndex ) {
    burnIndex = BigInt(burnIndex)
	let lastIndex = totalSupply()
    lastIndex = lastIndex-1n

	Mev.SetContractData(makeNFTIndexKey(nftID), "")

	if (burnIndex != lastIndex) {
		let lastNFTID = tokenByIndex(lastIndex)
		setNFTIndex(lastNFTID, burnIndex)
	}
	Mev.SetContractData(makeIndexNFTKey(lastIndex), "")
	Mev.SetContractData(tagNFTCount, lastIndex)
}

function removeNFTAccount(nftID) {
	let addr = ownerOf(nftID)

	let indexbs = Mev.AccountData(addr, makeNFTIndexKey(nftID))
    require(indexbs != "", "not exist nft")
	let burnIndex = BigInt(indexbs)

	let lastIndex = balanceOf(addr)
    lastIndex = lastIndex - 1n

	Mev.SetAccountData(addr, makeNFTIndexKey(nftID), "")

    if (burnIndex != lastIndex) {
		let lastNFTID = tokenOfOwnerByIndex(addr, lastIndex)
		setNFTAccountIndex(addr, lastNFTID, burnIndex)
	}
	Mev.SetAccountData(addr, makeIndexNFTKey(lastIndex), "")
	Mev.SetAccountData(addr, tagNFTCount, lastIndex)
}

function setBaseURI(uri) {
    require(isOwner(), "doesn't have setBaseURI permission")
	Mev.SetContractData(tagBaseURI, uri)
}

function setTokenURI(tokenID, uri) {
    require(isOwner(), "doesn't have setTokenURI permission")
	Mev.SetContractData(makeTokenURIKey(tokenID), uri)
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
    checkTransferPermission(_tokenId, _from, _to)
    removeNFTAccount(_tokenId)
    addNFTAccount(_to, _tokenId)
}

function checkTransferPermission(_tokenId, _from, _to) {
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
	Mev.SetContractData(makeTokenApproveKey(_tokenId), address(_approved))
}

/// @notice Enable or disable approval for a third party ("operator") to manage
///  all of `msg.sender`'s assets
/// @dev Emits the ApprovalForAll event. The contract MUST allow
///  multiple operators per owner.
/// @param _operator Address to add to the set of authorized operators
/// @param _approved True if the operator is approved, false to revoke approval
function setApprovalForAll(_operator, _approved) {
	if (_approved === true || _approved == "true" ) {
        let key = makeTokenApproveForAllKey(Mev.From(), _operator)
		Mev.SetContractData(key, 1)
	} else {
		Mev.SetContractData(makeTokenApproveForAllKey(Mev.From(), _operator), "")
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

function makeNFTKey(key, body) {
	return key+""+BigInt(body).toString(16)
}

function makeNFTIndexKey(k) {
    k = BigInt(k)
	return makeNFTKey(tagNFTIndex, k)
}
function makeIndexNFTKey(i) {
	return makeNFTKey(tagIndexNFT, BigInt(i))
}
function makeNFTOwnerKey(tokenID) {
	return makeNFTKey(tagNFTOwner, BigInt(tokenID))
}
function makeTokenApproveKey(i) {
    i = BigInt(i)
	return makeNFTKey(tagTokenApprove, i)
}
function makeTokenApproveForAllKey(_owner, _operator) {
	return makeNFTKey(tagTokenApproveForAll, BigInt(_owner)*10000000000000000000000000000000000000000n + BigInt(_operator))
}
function makeTokenURIKey(tokenID) {
	tokenID = BigInt(tokenID)
	return makeNFTKey(tagTokenURI, tokenID)
}
