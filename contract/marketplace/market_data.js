const OWNER                            = "01"
const MANAGER                          = "02"
const MARKET                           = "03"
const MARKETOPERATIONCONTRACT          = "04"
const FEEBALANCEKEY                    = "05"
const SUGGESTEDPRICELISTKEY            = "06"
const MARKETCOLLECTIONITEMSKEY         = "07"
const FOUNDATIONADMINWALLETADDRESSKEY  = "08"
const ISLISTINGINMARKETKEY             = "09"
const FOUNDATIONBALANCESKEY            = "10"
const ERC20CONTRACTSKEY                = "11"
const BURNWALLETADDRESSKEY             = "12"

// const zeroAddress = "0x0000000000000000000000000000000000000000"

const SaleType = {
    BUY_NOW: "BUY_NOW",
    BIDDING: "BIDDING",
}

const State =  {
    OPEN:"OPEN",
    COMPLETED:"COMPLETED",
    CANCELED:"CANCELED",
    CLOSED_BY_MANAGER:"CLOSED_BY_MANAGER",
}

const CurrencyType =  {
    MEV: "MEV",
    USDT: "USDT",
}

function NftItem (serial) {
    let obj = {};
    try {
        obj = JSON.parse(serial)
    } catch(e) {}
    this.TypeOf = "NftItem"

    this.tokenContract = obj.tokenContract?obj.tokenContract:"";
    this.tokenId = obj.tokenId?obj.tokenId:"";
}

function SuggestItem (serial) {
    let obj = {};
    try {
        obj = JSON.parse(serial)
    } catch(e) {}
    this.TypeOf = "SuggestItem"

    this.buyer = obj.buyer?address(obj.buyer):"";
    this.seller = obj.seller?address(obj.seller):"";
    this.price = obj.price?BigInt(obj.price):BigInt(0);
}

function MarketItemData(serial) {
    let obj = {};
    try {
        obj = JSON.parse(serial)
    } catch(e) {}
    this.TypeOf = "MarketItemData"

    this.seller = obj.seller?address(obj.seller):"";
    this.buyer = obj.buyer?address(obj.buyer):"";
    if (obj.nft) {
        if (typeof obj.nft == "string") {
            this.nft = new NftItem(obj.nft)
        } else {
            this.nft = obj.nft
        }
    }
    this.buyNowPrice = obj.buyNowPrice?BigInt(obj.buyNowPrice):BigInt(0);
    this.currency = obj.currency?obj.currency:"";
    this.createTimeUtc = obj.createTimeUtc?obj.createTimeUtc:"";
    this.openTimeUtc = obj.openTimeUtc?obj.openTimeUtc:"";
    this.closeTimeUtc = obj.closeTimeUtc?obj.closeTimeUtc:"";
    this.state = obj.state?obj.state:"";
}

function Init(owner) {
    if (typeof owner === "undefined") {
        throw "owner not provided"
    }
	Mev.SetContractData(OWNER, address(owner))
	Mev.SetContractData(MANAGER, address(owner))
}

function onlyMarket() {
    let com = Mev.ContractData(MARKETOPERATIONCONTRACT)
    if (address(Mev.From()) != com) {
        throw `onlyMarket ${Mev.From()} ${com}`
    }
}

function onlyManager() {
    let com = Mev.ContractData(MANAGER)
    if (address(Mev.From()) != com) {
        throw `onlyManager ${Mev.From()} ${com}`
    }
}

function onlyOwner() {
    let com = Mev.ContractData(OWNER)
    if (address(Mev.From()) != com) {
        throw `onlyOwner ${Mev.From()} ${com}`
    }
}

// address private _marketOperationContract;
function _marketOperationContract() {
    return Mev.ContractData(MARKETOPERATIONCONTRACT)
}
function _setMarketOperationContract(addr) {
    return Mev.SetContractData(MARKETOPERATIONCONTRACT, address(addr))
}

// mapping(CurrencyType => uint256) public _feeBalance;
function _feeBalanceKey(CurrencyType) {
    return FEEBALANCEKEY+CurrencyType
}
function _feeBalance(CurrencyType) {
    return BigInt(Mev.ContractData(_feeBalanceKey(CurrencyType)))
}
function _setFeeBalance(CurrencyType, val) {
    val = BigInt(val)
    return Mev.SetContractData(_feeBalanceKey(CurrencyType), val.toString())
}

// mapping(address => mapping(uint256 => mapping(CurrencyType => SuggestItem[]))) private _suggestedPriceList;
function _suggestedPriceListKey(addr, tokenId, currency) {
    addr = address(addr)
    tokenId = BigInt(tokenId)
    return SUGGESTEDPRICELISTKEY+addr+tokenId+":"+currency
}
function _suggestedPriceList(addr, tokenId, currency) {
    addr = address(addr)
    tokenId = BigInt(tokenId)
    let list = Mev.ContractData(_suggestedPriceListKey(addr, tokenId, currency))
    if (list == "") {
        return "[]"
    }
    return list
}
function _setSuggestedPriceList(addr, tokenId, currency, list) {
    require(Array.isArray(list), "list is not array")
    return Mev.SetContractData(_suggestedPriceListKey(addr, tokenId, currency), JSON.stringify(list))
}

// mapping(address => uint256) private _itemCount;
function _itemCountKey(addr) {
    return ITEMCOUNT+address(addr)
}
function _itemCount(addr) {
    return BigInt(Mev.ContractData(_itemCountKey(addr)))
}
function _setItemCount(addr, i) {
    i = BigInt(i)
    Mev.SetContractData(_itemCountKey(addr), i.toString())
}

// mapping(address => mapping(uint256 => MarketItemData)) private _marketCollectionItems;
function _marketCollectionItemsKey(addr, tokenID) {
    tokenID = BigInt(tokenID)
    return MARKETCOLLECTIONITEMSKEY+address(addr)+tokenID
}
function _marketCollectionItems(addr, tokenID) {
    let raw = Mev.ContractData(_marketCollectionItemsKey(addr, tokenID))
    return new MarketItemData(raw)
}
function _setMarketCollectionItems(addr, tokenID, mid) {
    require(mid.TypeOf == "MarketItemData", "MarketItemData is not obj")
    Mev.SetContractData(_marketCollectionItemsKey(addr, tokenID), JSON.stringify(mid))
}

// mapping(address => address) private _foundationAdminWalletAddress;
function _foundationAdminWalletAddressKey(addr) {
    return FOUNDATIONADMINWALLETADDRESSKEY+address(addr)
}
function _foundationAdminWalletAddress(addr) {
    return Mev.ContractData(_foundationAdminWalletAddressKey(addr))
}
function _setFoundationAdminWalletAddress(addr, waddr) {
    Mev.SetContractData(_foundationAdminWalletAddressKey(addr), address(waddr))
}

// mapping(address => address) private _burnWalletAddress;
function _burnWalletAddressKey(addr) {
    return BURNWALLETADDRESSKEY+address(addr)
}
function _burnWalletAddress(addr) {
    return Mev.ContractData(_burnWalletAddressKey(addr))
}
function _setBurnWalletAddress(addr, waddr) {
    Mev.SetContractData(_burnWalletAddressKey(addr), address(waddr))
}
// mapping(address => mapping(uint256 => bool)) private _isListingInMarket; // track if item if in active market
function _isListingInMarketKey(addr, tokenID) {
    return ISLISTINGINMARKETKEY+address(addr)+tokenID
}
function _isListingInMarket(addr, tokenID) {
    return Mev.ContractData(_isListingInMarketKey(addr, tokenID)) == "true"
}
function _setIsListingInMarket(addr, tokenID, val) {
    Mev.SetContractData(_isListingInMarketKey(addr, tokenID), val?true:false)
}
// mapping(address => mapping(CurrencyType => uint256)) private _foundationBalances;
function _foundationBalancesKey(addr, CurrencyType) {
    return FOUNDATIONBALANCESKEY+address(addr)+CurrencyType
}
function _foundationBalances(addr, CurrencyType) {
    return Mev.ContractData(_foundationBalancesKey(addr, CurrencyType))
}
function _setFoundationBalances(addr, CurrencyType, amt) {
    amt = BigInt(amt)
    Mev.SetContractData(_foundationBalancesKey(addr, CurrencyType), amt)
}
// mapping(CurrencyType => IERC20) public _erc20Contracts;
function _erc20ContractsKey(CurrencyType) {
    return ERC20CONTRACTSKEY+CurrencyType
}
function _erc20Contracts(CurrencyType) {
    let addr = Mev.ContractData(_erc20ContractsKey(CurrencyType))
    if (addr == "") {
        throw "not supported currency("+CurrencyType+")"
    }
    return address(addr)
}
function _setErc20Contracts(CurrencyType, addr) {
    Mev.SetContractData(_erc20ContractsKey(CurrencyType), address(addr))
}

function getMarketItem(token, tokenId) {
    return _marketCollectionItems(token, tokenId);
}

function getERC20Contract(currency) {
    return _erc20Contracts(currency);
}

function getFoundationAdminAddress(token) {
    return _foundationAdminWalletAddress(token);
}

function getBurnAddress(token) {
    return _burnWalletAddress(token);
}

function getFoundationRoyalty(wallet, currency) {
    return _foundationBalances(wallet, currency);
}

function totalMarketCollectionItems(token) {
    return _itemCount(token);
}

function isTokenRegistered(token, tokenId) {
    return  _isListingInMarket(token, tokenId);
}

function getMarketOperationAddress() {
    return _marketOperationContract();
}

function getItemSuggestionInfo(nftAddress, tokenId, currency) {
    onlyMarket()
    let list = _suggestedPriceList(nftAddress, tokenId, currency);
    list = JSON.parse(list)

    let len = list.length;
    let item
    if (len > 0) {
        item = list[len - 1];
    }
    return item;
}

function getItemSuggestionInfos(nftAddress, tokenId, currency) {
    onlyMarket()
    nftAddress = address(nftAddress)
    tokenId = BigInt(tokenId)
    return _suggestedPriceList(nftAddress, tokenId, currency);
}

function setERC20Contract(currency, token) {
    onlyOwner()
    _setErc20Contracts(currency, token);
}

/**
 * 리스트에 해당 값을 추가한다.
 */
function suggestItemToBuy(suggester, seller, nftAddress, tokenId, suggestBiddingPrice, currency) {
    onlyMarket()
    nftAddress = address(nftAddress)
    tokenId = BigInt(tokenId)
    let list = _suggestedPriceList(nftAddress, tokenId, currency);
    list = JSON.parse(list)

    let item = new SuggestItem();
    item.buyer = suggester
    item.seller = seller
    item.price = BigInt(suggestBiddingPrice)

    list.push(item);
    _setSuggestedPriceList(nftAddress, tokenId, currency, list);

    // emit MarketItemSuggested(suggester, nftAddress, tokenId, suggestBiddingPrice, currency);
}

function cancelItemToBuy(suggester, nftAddress, tokenId, currency, suggestBiddingPrice) {
    onlyMarket()
    suggester = address(suggester)
    suggestBiddingPrice = BigInt(suggestBiddingPrice)
    tokenId = BigInt(tokenId)

    let list = _suggestedPriceList(nftAddress, tokenId, currency);
    list = JSON.parse(list)
    let len = list.length;
    let index = 0;
    for (let i = 0; i < len; i ++) {
        if (address(list[i].buyer) == address(suggester) && BigInt(list[i].price) == BigInt(suggestBiddingPrice)) {
            index = i;
            if (index >= 0) {
                _removeArray(nftAddress, tokenId, currency, index);
                // emit MarketItemSuggestionCanceled(suggester, nftAddress, tokenId, suggestBiddingPrice);
                return
            }
        }
    }

    throw "not exist biddingPrice"
}

function acceptItemSuggestion(seller, nftAddress, tokenId, suggestBiddingPrice, currency)  {
    onlyMarket()
    seller = address(seller)
    suggestBiddingPrice = BigInt(suggestBiddingPrice)
    tokenId = BigInt(tokenId)

    let list = _suggestedPriceList(nftAddress, tokenId, currency);
    list = JSON.parse(list)
    
    let len = list.length;
    let selectedItem;
    for (let i = 0; i < len; i ++) {
        if (list[i].price == suggestBiddingPrice && list[i].seller == seller) {
            selectedItem = list[i];
            break;
        }
    }
    require(selectedItem.price == suggestBiddingPrice, "Invaild suggestBiddingPrice price");

    let marketItem = _marketCollectionItems(nftAddress, tokenId);
    if (marketItem && marketItem.nft && 
        address(marketItem.nft.tokenContract) == address(nftAddress) &&
        BigInt(marketItem.nft.tokenId) == BigInt(tokenId)
        ) {
            _checkMarketItem(marketItem);
            marketItem.state = State.CANCELED
            _setMarketCollectionItems(nftAddress, tokenId, marketItem);
            _setIsListingInMarket(nftAddress, tokenId);
    }

    _setSuggestedPriceList(nftAddress, tokenId, currency, []);
    for (let _currency in CurrencyType) {
        if (currency != _currency) {
            _setSuggestedPriceList(nftAddress, tokenId, _currency, []);
        }
    }
    // emit TransactCompletedItemInMarket(nftAddress, tokenId, selectedItem.buyer, selectedItem.price, State.COMPLETED);
}

function _checkMarketItem(item) {
    require(item.TypeOf == "MarketItemData", "item is not MarketItemData");
    require(item.seller != address(0), "Invalid item - seller" );
    require(address(item.nft.tokenContract) != address(0), "Invalid item - nft token contract" );
}

function registerItemInMarket(item) {
    onlyMarket()
    item = new MarketItemData(item)
    require(item.TypeOf == "MarketItemData", "item is not MarketItemData");
    _checkMarketItem(item);
  
    let nftAddress = address(item.nft.tokenContract);
    let tokenId = BigInt(item.nft.tokenId);

    _setMarketCollectionItems(nftAddress, tokenId, item);
    _setIsListingInMarket(nftAddress, tokenId, true);

    // emit MarketItemRegistered(nftAddress, tokenId, item);
}

function cancelItemInMarket(nftAddress, tokenId, state) {
    onlyMarket()
    let marketItem = _marketCollectionItems(nftAddress, tokenId);
    _checkMarketItem(marketItem);
    
    marketItem.state = state
    _setMarketCollectionItems(nftAddress, tokenId, marketItem);
    _setIsListingInMarket(nftAddress, tokenId);
    
    // emit MarketItemCanceled(nftAddress, tokenId, marketItem, state);
}

function transactCompleteItemInMarket(item) {
    onlyMarket()
    if (typeof item == "string") {
        item = new MarketItemData(item)
    }
    require(item.TypeOf == "MarketItemData", "item is not MarketItemData");

    _checkMarketItem(item);
    
    item = _marketCollectionItems(item.nft.tokenContract, item.nft.tokenId)
    item.state = State.COMPLETED;
    _setMarketCollectionItems(item.nft.tokenContract, item.nft.tokenId, item)

    _setIsListingInMarket(item.nft.tokenContract, item.nft.tokenId);
    for (let currency in CurrencyType) {
        _setSuggestedPriceList(item.nft.tokenContract, item.nft.tokenId, currency, []);
    }
    // emit TransactCompletedItemInMarket(address(item.nft.tokenContract), item.nft.tokenId, item.buyer, item.buyNowPrice, State.COMPLETED);
}

function addFoundationRoyalty(foundationAddress, amount, currency) {
    onlyMarket()
    foundationAddress = address(foundationAddress)
    amount = BigInt(amount)
    require( foundationAddress != address(0), "Invaild address");
    require( 0 < amount, "Zero amount");
    

    let amt = _foundationBalances(foundationAddress, currency)

    _setFoundationBalances(foundationAddress, currency, amt+amount);
    // emit FoundationRoyaltyAdded(foundationAddress, amount, currency);
}

function addMarketFee(amount, currency) {
    onlyMarket()
    amount = BigInt(amount)
    require( 0 < amount, "Zero amount");
    let fee = _feeBalance(currency) 
    _setFeeBalance(currency, fee + amount)
    // emit MarketFeeAdded(amount, currency);
}

function setBuyer(nftAddress, tokenId, buyer) {
    onlyMarket()
    let item = _marketCollectionItems(nftAddress, tokenId);
    item.buyer = buyer
    _setMarketCollectionItems(nftAddress, tokenId, item);
}

// Market Fees
function approveFeesForERC20Token(spender, currency) {
    onlyMarket()
    spender = address(spender)
    require(spender != address(0), "Invaild spender");
    let erc20Addr = _erc20Contracts(currency)
    require( erc20Addr != address(0), "Invalid Currency" );

    let amount = _feeBalance(currency);
    Mev.Exec(erc20Addr, "approve", [spender, amount])

    // emit FeesForERC20TokenApproved(spender, amount, currency);
    return amount;
}

function setFoundationAdminAddress(nftAddress, adminAddress) {
    onlyOwner()
    nftAddress = address(nftAddress)
    adminAddress = address(adminAddress)
    require( nftAddress != address(0), "Invaild NFT address");
    require( adminAddress != address(0), "Invaild admin address");
    _setFoundationAdminWalletAddress(nftAddress, adminAddress);
    // emit ChangedFoundationAdminAddress(nftAddress, adminAddress);
}

function setBurnAddress(nftAddress, burnAddress) {
    onlyOwner()
    nftAddress = address(nftAddress)
    burnAddress = address(burnAddress)
    require( nftAddress != address(0), "Invaild NFT address");
    require( burnAddress != address(0), "Invaild admin address");
    _setBurnWalletAddress(nftAddress, burnAddress);
    // emit ChangedBurnAddress(nftAddress, burnAddress);
}

function setMandatoryInitContract(market) {
    onlyOwner()
    market = address(market)
    require( market != address(0), "Invaild market address");
    _setMarketOperationContract(market);
    // emit MandatoryInitCompleted(market);
}

// gather fees
function collectFees(to, currency) {
    onlyMarket()
    to = address(to)
    let amount = _feeBalance(currency);
    require(0 < amount, "NftMarket.collectFees: NO_FEES");

    _setFeeBalance(currency, 0n);
    _transfer(to, amount, currency);

    // emit MarketFeesCollected(to, amount, currency);
    return amount;
}

function _transfer(to, amount, currency) {
    to = address(to)
    amount = BigInt(amount)
    let erc20addr = _erc20Contracts(currency)
    Mev.Exec(erc20addr , "transfer", [to, amount])
}

function _removeArray(nftAddress, tokenId, currency, index) {
    let list = _suggestedPriceList(nftAddress, tokenId, currency);
    list = JSON.parse(list)
    list.splice(index,1);
    _setSuggestedPriceList(nftAddress, tokenId, currency, list)
}
