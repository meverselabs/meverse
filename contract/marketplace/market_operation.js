const OWNER                     = "01"
const MANAGER                   = "02"

const MARKET_FEE                = "03"
const ROYALTY_FEE               = "04"
const MARKETDATA                = "05"
const BURN_FEE                  = "06"

const MAX_PERCENT = 10000n; // 2 decimal precision for percentage (100.00%)
const MAX_PRICE = 1157920892373161954235709850086879078532699846656405640394575840079131296n; // = ((2^256)-1)/10000, since to avoid multiplication overflow we should satisfy X*10000<=(2^256-1)

function Init(owner, marketFeeStr, royaltyFeeStr) {
    
    if (typeof owner === "undefined") {
        throw "owner not provided"
    }
    let marketFee = BigInt(marketFeeStr);
    require(marketFee <= MAX_PERCENT , "INVALID_BUYNOW_FEE");
    let royaltyFee = BigInt(royaltyFeeStr);
    require(royaltyFee <= MAX_PERCENT , "INVALID_ROYALTY_FEE");

    // _MARKET_FEE_ = marketFee;
    // _ROYALTY_FEE_ = royaltyFee;
    
    owner = address(owner)
	Mev.SetContractData(OWNER, owner)
	Mev.SetContractData(MANAGER, owner)
    _setMarketFee(marketFee)
    _setRoyaltyFee(royaltyFee)
}

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
    this.tokenId = obj.tokenId?BigInt(obj.tokenId):"";
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

function onlyManager() {
    if (!isManager()) {
        throw `onlyManager ${Mev.From()} ${com}`
    }
}
function isManager() {
    let com = Mev.ContractData(MANAGER)
    return address(Mev.From()) == com 
}

function onlyOwner() {
    let com = Mev.ContractData(OWNER)
    if (address(Mev.From()) != com) {
        throw `onlyOwner ${Mev.From()} ${com}`
    }
}

function setOwner(addr) {
    onlyOwner()
    addr = address(addr)
    Mev.SetContractData(OWNER, addr)
}

function setManager(addr) {
    onlyOwner()
    addr = address(addr)
    Mev.SetContractData(MANAGER, addr)
}

// NftMarketData _marketData;
function _marketData() {
    return Mev.ContractData(MARKETDATA)
}
function _setMarketData(addr) {
    return Mev.SetContractData(MARKETDATA, address(addr))
}

// uint256 private _MARKET_FEE_;
function _marketFee() {
    return BigInt(Mev.ContractData(MARKET_FEE))
}
function _setMarketFee(val) {
    val = BigInt(val)
    return Mev.SetContractData(MARKET_FEE, val)
}
// uint256 private _ROYALTY_FEE_;
function _royaltyFee() {
    return BigInt(Mev.ContractData(ROYALTY_FEE))
}
function _setRoyaltyFee(val) {
    val = BigInt(val)
    return Mev.SetContractData(ROYALTY_FEE, val)
}

function _burnFee() {
    return BigInt(Mev.ContractData(BURN_FEE))
}
function _setBurnFee(val) {
    val = BigInt(val)
    return Mev.SetContractData(BURN_FEE, val)
}

function transferFrom(nftAddress, owner, to, tokenId) {
    require(address(owner) == address(Mev.From()), "not owner")
    Mev.Exec(nftAddress, "transferFrom", [owner, to, tokenId]);
}

function getItemSuggestionInfos(nftAddress, tokenId, currency) {
    // return _marketData.getItemSuggestionInfos(nftAddress, tokenId, currency);
    nftAddress = address(nftAddress)
    tokenId = BigInt(tokenId)
    return Mev.Exec(_marketData(), "getItemSuggestionInfos", [nftAddress, tokenId, currency])
}

function getMarketItem(token, tokenId) {
    return Mev.Exec(_marketData(), "getMarketItem", [token, tokenId])
}

function getMarketDataContractAddress() {
    return _marketData();
}

function getMarketFee() {
    return _marketFee();
}

function getRoyaltyFee() {
    return _royaltyFee()
}

function suggestItemToBuyWithSuggester(nftAddress, tokenId, suggestBiddingPrice, currency) {
    nftAddress = address(nftAddress)
    tokenId = BigInt(tokenId)

    suggestBiddingPrice = BigInt(suggestBiddingPrice)
    require(nftAddress != address(0), "NftMarket.suggestItemToBuy: Invaild token contract");
    
    // let item = _marketData.getItemSuggestionInfo(nftAddress, tokenId, currency);
    let item = Mev.Exec(_marketData(), "getItemSuggestionInfo", [nftAddress, tokenId, currency])
    item = new SuggestItem(item)

    require(suggestBiddingPrice > item.price,"NftMarket.suggestItemToBuy: not valid suggestBiddingPrice");
    // IERC20 erc20Contract = _marketData.getERC20Contract(currency);
    let erc20Contract = Mev.Exec(_marketData(), "getERC20Contract", [currency])
    erc20Contract = address(erc20Contract)
    require(erc20Contract != address(0), "NftMarket.suggestItemToBuy: Currency contract not registered currency:"+currency);
    let allowance = Mev.Exec(erc20Contract, "allowance", [_msgSender(), address(Mev.This())])
    allowance = BigInt(allowance)
    require(suggestBiddingPrice <= allowance, "NftMarket.suggestItemToBuy: is not allowanced");
    let balanceOf = Mev.Exec(erc20Contract, "balanceOf", [_msgSender()])
    balanceOf = BigInt(balanceOf)
    require(suggestBiddingPrice <= balanceOf, "NftMarket.suggestItemToBuy: is not sufficient balance");
    
    // Inft nft = Inft(nftAddress);
    // address nftOwner = nft.ownerOf(tokenId);
    let nftOwner = Mev.Exec(nftAddress, "ownerOf", [tokenId])
    nftOwner = address(nftOwner)
    // _marketData.suggestItemToBuy(_msgSender(), nftOwner, nftAddress, tokenId, suggestBiddingPrice, currency);
    Mev.Exec(_marketData(), "suggestItemToBuy", [_msgSender(), nftOwner, nftAddress, tokenId, suggestBiddingPrice, currency])
    // emit MarketItemSuggested(_msgSender(), nftAddress, tokenId, suggestBiddingPrice, currency);
}

function cancelItemToBuyWithSuggester(nftAddress, tokenId, currency, suggestBiddingPrice) {
    // _marketData.cancelItemToBuy(_msgSender(), nftAddress, tokenId, currency, suggestBiddingPrice);
    nftAddress = address(nftAddress)
    suggestBiddingPrice = BigInt(suggestBiddingPrice)
    
    let erc20Contract = Mev.Exec(_marketData(), "getERC20Contract", [currency])
    erc20Contract = address(erc20Contract)
    require(erc20Contract != address(0), "NftMarket.registerMarketItem: NOT_SUPPORTED_CURRECY");


    Mev.Exec(_marketData(), "cancelItemToBuy", [_msgSender(), nftAddress, tokenId, currency, suggestBiddingPrice])
    // emit MarketItemSuggestionCanceled(_msgSender(), nftAddress, tokenId, suggestBiddingPrice);
}

function acceptItemToBuyWithSeller(nftAddress, tokenId, suggestedBiddingPrice, currency) {
    nftAddress = address(nftAddress);
    tokenId = BigInt(tokenId)

    let erc20Contract = Mev.Exec(_marketData(), "getERC20Contract", [currency])
    erc20Contract = address(erc20Contract)
    require(erc20Contract != address(0), "NftMarket.registerMarketItem: NOT_SUPPORTED_CURRECY");

    {
        let isApp = Mev.Exec(nftAddress, "isApprovedForAll", [_msgSender(), Mev.This()])
        require(isApp == "true", "NftMarket.acceptItemToBuy: NOT_APPROVED_OR_INVALID_TOKEN_ID");
        // address nftOwner = nft.ownerOf(tokenId);
        let nftOwner = Mev.Exec(nftAddress, "ownerOf", [tokenId])
        nftOwner = address(nftOwner)
        require(nftOwner == _msgSender(), "NftMarket.acceptItemToBuy: seller is not owner");
    }
    suggestedBiddingPrice = BigInt(suggestedBiddingPrice)
    // SuggestItem memory suggestItem = _marketData.getItemSuggestionInfo(nftAddress, tokenId, currency);
    let siStr = Mev.Exec(_marketData(), "getItemSuggestionInfos", [nftAddress, tokenId, currency]);
    let rawSi = JSON.parse(siStr)
    let suggestItem = {};
    for (var i in rawSi) {
        let si = JSON.parse(rawSi[i])
        if (suggestedBiddingPrice == BigInt(si.price)) {
            suggestItem = new SuggestItem(JSON.stringify(si));
        }
    }
    require(suggestItem.TypeOf == "SuggestItem", "NftMarket.acceptItemToBuy: Invalid suggestItem");
    require(suggestItem.price == suggestedBiddingPrice, "NftMarket.acceptItemToBuy: Invalid price");
    // address adminAddress = _marketData.getFoundationAdminAddress(nftAddress);
    let adminAddress = Mev.Exec(_marketData(), "getFoundationAdminAddress", [nftAddress]);
    adminAddress = address(adminAddress)
    require(adminAddress != address(0), "NftMarket.acceptItemToBuy: foundation admin is not setted");
    let burnAddress = Mev.Exec(_marketData(), "getBurnAddress", [nftAddress]);
    burnAddress = address(burnAddress)
    require(burnAddress != address(0), "NftMarket.acceptItemToBuy: burn is not setted");
    
    let totalAmountForMarket = 0n;
    let remainAmount = suggestedBiddingPrice;
    let marketFee = _calcMarketFee(remainAmount);
    let foundationRoyalty = _calcRoyaltyFee(remainAmount);
    let burnCap = _calcBurnFee(remainAmount);
    {
        remainAmount = remainAmount - marketFee - foundationRoyalty - burnCap;
        totalAmountForMarket = totalAmountForMarket + marketFee + foundationRoyalty + burnCap;
        // _marketData.addMarketFee(marketFee, currency);
        Mev.Exec(_marketData(), "addMarketFee", [marketFee, currency]);
        // _marketData.addFoundationRoyalty(adminAddress, foundationRoyalty, currency);
        Mev.Exec(_marketData(), "addFoundationRoyalty", [adminAddress, foundationRoyalty, currency]);
        //transfer NFT to buyer
        // IERC721(nftAddress).transferFrom(_msgSender(), suggestItem.buyer, tokenId);
        Mev.Exec(nftAddress, "transferFrom", [_msgSender(), suggestItem.buyer, tokenId]);
        require(totalAmountForMarket+remainAmount == suggestedBiddingPrice, "NftMarket.acceptItemToBuy: Not match amount and remainAmount + totalAmountForMarket");
    }
    {
        // IERC20 erc20Contract = _marketData.getERC20Contract(currency);
        let erc20Contract = Mev.Exec(_marketData(), "getERC20Contract", [currency]);
        erc20Contract = address(erc20Contract)
        require(erc20Contract != address(0), "NftMarket.acceptItemToBuy: Currency contract not registered");
        // erc20Contract.allowance(suggestItem.buyer, address(this))
        let erc20allow = Mev.Exec(erc20Contract, "allowance", [suggestItem.buyer, Mev.This()]);
        erc20allow = BigInt(erc20allow)
        require(suggestedBiddingPrice <= erc20allow, "NftMarket.acceptItemToBuy: is not allowanced");
        
        // erc20Contract.transferFrom(suggestItem.buyer, address(_marketData), marketFee);
        // erc20Contract.transferFrom(suggestItem.buyer, address(adminAddress), foundationRoyalty);
        // erc20Contract.transferFrom(suggestItem.buyer, _msgSender(), remainAmount);
        Mev.Exec(erc20Contract, "transferFrom", [suggestItem.buyer, _marketData(), marketFee]);
        Mev.Exec(erc20Contract, "transferFrom", [suggestItem.buyer, adminAddress, foundationRoyalty]);
        Mev.Exec(erc20Contract, "transferFrom", [suggestItem.buyer, burnAddress, burnCap]);
        Mev.Exec(erc20Contract, "transferFrom", [suggestItem.buyer, _msgSender(), remainAmount]);
        // _marketData.acceptItemSuggestion(_msgSender(), nftAddress, tokenId, suggestedBiddingPrice, currency);
        Mev.Exec(_marketData(), "acceptItemSuggestion", [_msgSender(), nftAddress, tokenId, suggestedBiddingPrice, currency]);
        // emit MarketItemCompleted(SaleType.BIDDING, nftAddress, tokenId, _msgSender(), suggestItem.buyer, address(erc20Contract), suggestedBiddingPrice);
    }
}

function registerMarketItem(nftAddress, tokenId, buyNowPrice, currency, openTimeUtc, closeTimeUtc) {
    nftAddress = address(nftAddress)
    // address owner = nftAddress.ownerOf(tokenId);
    let owner = Mev.Exec(nftAddress, "ownerOf", [BigInt(tokenId)]);
    owner = address(owner)
    require(owner == _msgSender(), "NftMarket.registerMarketItem: NOT_OWNER");
    let isapp = Mev.Exec(nftAddress, "isApprovedForAll", [owner, Mev.This()]);
    require(isapp == "true", "NftMarket.registerMarketItem: NOT_APPROVED_OR_INVALID_TOKEN_ID");
    buyNowPrice = BigInt(buyNowPrice)
    require(0n < buyNowPrice, "NftMarket.registerMarketItem: ZERO_PRICE");
    require(buyNowPrice <= MAX_PRICE, "NftMarket.registerMarketItem: INVALID_PRICE");
    // bool isRegistered = _marketData.isTokenRegistered(address(nftAddress), tokenId);
    let isRegistered = Mev.Exec(_marketData(), "isTokenRegistered", [nftAddress, tokenId]);
    isRegistered = isRegistered=="true"
    
    let erc20Contract = Mev.Exec(_marketData(), "getERC20Contract", [currency])
    erc20Contract = address(erc20Contract)
    require(erc20Contract != address(0), "NftMarket.registerMarketItem: NOT_SUPPORTED_CURRECY");

    // MarketItemData memory marketItem = _marketData.getMarketItem(address(nftAddress), tokenId);
    let marketItem = Mev.Exec(_marketData(), "getMarketItem", [nftAddress, tokenId]);
    marketItem = new MarketItemData(marketItem)
    if (isRegistered) {
        require(marketItem.seller != owner, "NftMarket.registerMarketItem: ALREADY_REGISTERED");
    }
    
    // NftItem memory nftItem = NftItem(nftAddress, tokenId);
    let nftItem = new NftItem()
    nftItem.tokenContract = nftAddress;
    nftItem.tokenId = tokenId;

    let item = new MarketItemData();
    item.seller =  _msgSender()
    item.buyer =  address(0)
    item.nft =  nftItem
    item.buyNowPrice =  buyNowPrice
    item.currency =  currency
    item.createTimeUtc = Mev.TargetHeight()
    item.openTimeUtc =  openTimeUtc
    item.closeTimeUtc =  closeTimeUtc
    item.state =  State.OPEN

    // _marketData.registerItemInMarket(item);
    Mev.Exec(_marketData(), "registerItemInMarket", [item]);
    // emit MarketItemRegistered(address(nftAddress), tokenId, item);
}

function cancelMarketItem(nftAddress, tokenId) {
    nftAddress = address(nftAddress)
    // MarketItemData memory marketItem = _marketData.getMarketItem(nftAddress, tokenId);
    let marketItem = Mev.Exec(_marketData(), "getMarketItem", [nftAddress, tokenId]);
    marketItem = new MarketItemData(marketItem)

    require(_msgSender() == marketItem.seller || isManager(_msgSender()), "NftMarket.cancelMarketItem: INVALID_CALLER");
    let itemState;
    if ( isManager(_msgSender()) ) {
        itemState = State.CLOSED_BY_MANAGER;
    } else {
        itemState = State.CANCELED;
    }

    // _marketData.cancelItemInMarket(nftAddress, tokenId, itemState);
    Mev.Exec(_marketData(), "cancelItemInMarket", [nftAddress, tokenId, itemState]);
    // emit MarketItemCanceled(nftAddress, tokenId, marketItem, itemState);
}

function buyNowWithToken(nftAddress, tokenId, amount, currency) {
    _buyNow(nftAddress, tokenId, amount, currency);
}

function setMarketFee(newFee) {
    onlyManager()
    newFee = BigInt(newFee)
    require(newFee <= MAX_PERCENT , "Invaild fee");
    // let oldFee = _marketFee();
    _setMarketFee(newFee)
    // _MARKET_FEE_ = newFee;
    // emit MarketFeeUpdate(oldFee, newFee);
}

function setRoyaltyFee(newFee) {
    onlyManager()
    newFee = BigInt(newFee)
    require(newFee <= MAX_PERCENT , "Invaild fee");
    // uint256 oldFee = _ROYALTY_FEE_;
    _setRoyaltyFee(newFee)
    // _ROYALTY_FEE_ = newFee;
    // emit RoyaltyFeeUpdate(oldFee, newFee);
}

function setBurnFee(newFee) {
    onlyManager()
    newFee = BigInt(newFee)
    require(newFee <= MAX_PERCENT , "Invaild fee");
    // uint256 oldFee = _ROYALTY_FEE_;
    _setBurnFee(newFee)
    // _ROYALTY_FEE_ = newFee;
    // emit RoyaltyFeeUpdate(oldFee, newFee);
}

function setMandatoryMarketDataContract(marketData) {
    onlyOwner()
    marketData = address(marketData)
    require(marketData != address(0), "Invaild market data contract address");
    // _marketData = NftMarketData(marketData);
    _setMarketData(marketData)
    // emit MandatoryMarketDataContractCompleted(marketData);
}

function approveFeesForERC20Token(currency) {
    onlyManager()
    // uint256 amount = _marketData.approveFeesForERC20Token(_msgSender(), currency);
    Mev.Exec(_marketData(), "approveFeesForERC20Token", [_msgSender(), currency]);
    // emit FeesForERC20TokenApproved(_msgSender(), amount, currency);
}

function collectFees(currency) {
    onlyManager()
    // uint256 amount = _marketData.collectFees(_msgSender(), currency);
    Mev.Exec(_marketData(), "collectFees", [_msgSender(), currency]);
    // emit MarketFeesCollected(_msgSender(), amount, currency);
}

function _buyNow(nftAddress, tokenId, amount, currency) {
    nftAddress = address(nftAddress)
    amount = BigInt(amount)
    tokenId = BigInt(tokenId)
    // MarketItemData memory marketItem = _marketData.getMarketItem(nftAddress, tokenId);
    let marketItem = Mev.Exec(_marketData(), "getMarketItem", [nftAddress, tokenId]);
    marketItem = new MarketItemData(marketItem)

    let erc20Contract = Mev.Exec(_marketData(), "getERC20Contract", [currency])
    erc20Contract = address(erc20Contract)
    require(erc20Contract != address(0), "NftMarket.buyNow: NOT_SUPPORTED_CURRECY");

    require(marketItem.seller != address(0) && marketItem.state == State.OPEN , "NftMarket.buyNow: INVALID_AUCTION_STATE ("+marketItem.state+")"); 
    require(marketItem.currency == currency, "NftMarket.buyNow: INVALID_CURRENCY ("+marketItem.currency+") and ("+currency+")");
    require(_msgSender() != marketItem.seller, "NftMarket.buyNow: INVALID_BUYER ("+_msgSender()+") and ("+marketItem.seller+")");
    // require(block.timestamp < marketItem.closeTimeUtc, "NftMarket.buyNow: Item sale closed.");
    require(Mev.TargetHeight()-0 < marketItem.closeTimeUtc-0, "NftMarket.buyNow: Item sale closed." + typeof Mev.TargetHeight() + " : " + typeof marketItem.closeTimeUtc);

    require(0n < amount, "NftMarket.buyNow: ZERO_VALUE");
    require(marketItem.buyNowPrice == amount, "Price not matched");

    // actualAmount = amount - fee
    // _marketData.setBuyer(nftAddress, tokenId, _msgSender());
    Mev.Exec(_marketData(), "setBuyer", [nftAddress, tokenId, _msgSender()]);
    _settle(marketItem, _msgSender(), amount, currency);
}

function _settle(marketItem, buyerAddress, amount, currency) {
    // Mev.Log("_settle", marketItem, marketItem.nft.tokenContract, buyerAddress, amount, currency)
    // Inft nft = Inft(address(marketItem.nft.tokenContract));
    let nft = address(marketItem.nft.tokenContract)
    // address adminAddress = _marketData.getFoundationAdminAddress(address(marketItem.nft.tokenContract));
    let adminAddress = Mev.Exec(_marketData(), "getFoundationAdminAddress", [nft]);
    adminAddress = address(adminAddress)
    // address nftOwner = nft.ownerOf(marketItem.nft.tokenId);
    let nftOwner = Mev.Exec(nft, "ownerOf", [BigInt(marketItem.nft.tokenId)]);
    nftOwner = address(nftOwner)
    
    require(adminAddress != address(0), "NftMarket._settle: foundation admin is not setted");

    let burnAddress = Mev.Exec(_marketData(), "getBurnAddress", [nft]);
    burnAddress = address(burnAddress)
    Mev.Log(burnAddress)
    require(burnAddress != address(0), "NftMarket.acceptItemToBuy: burn is not setted");

    // require(nft.isApprovedForAll(marketItem.seller, address(this)), "NftMarket._settle: NOT_APPROVED_OR_INVALID_TOKEN_ID");
    let isapp = Mev.Exec(nft, "isApprovedForAll", [marketItem.seller, Mev.This()])
    require(isapp == "true", "NftMarket._settle: NOT_APPROVED_OR_INVALID_TOKEN_ID");
    require(nftOwner == marketItem.seller, "NftMarket._settle: seller is Not owner");
    
    amount = BigInt(amount)
    buyerAddress = address(buyerAddress)

    let remainAmount = amount;
    let marketFee = 0n;
    let foundationRoyalty = 0n;
    let burnCap = 0n;
    {
        let totalAmountForMarket = 0n;
        marketFee = _calcMarketFee(remainAmount);
        foundationRoyalty = _calcRoyaltyFee(remainAmount);
        burnCap = _calcBurnFee(remainAmount);
        remainAmount = remainAmount - marketFee - foundationRoyalty - burnCap;
        totalAmountForMarket = totalAmountForMarket + marketFee + foundationRoyalty + burnCap;
        // _marketData.addMarketFee(marketFee, currency);
        Mev.Exec(_marketData(), "addMarketFee", [marketFee, currency])
        
        // _marketData.addFoundationRoyalty(adminAddress, foundationRoyalty, currency);
        Mev.Exec(_marketData(), "addFoundationRoyalty", [adminAddress, foundationRoyalty, currency])
        
        // transfer NFT to buyer (actual transfer)
        // marketItem.nft.tokenContract.transferFrom(address(marketItem.seller), buyerAddress, marketItem.nft.tokenId);
        Mev.Exec(marketItem.nft.tokenContract, "transferFrom", [marketItem.seller, buyerAddress, BigInt(marketItem.nft.tokenId)])
        require(totalAmountForMarket + remainAmount == amount, "NftMarket._settle: Not match amount and remainAmount + totalAmountForMarket");
    }
    {
        // IERC20 erc20Contract = _marketData.getERC20Contract(currency);
        let erc20Contract = Mev.Exec(_marketData(), "getERC20Contract", [currency])
        erc20Contract = address(erc20Contract)
        require(erc20Contract != address(0), "NftMarket._settle: Currency contract not registered");
        // require(amount <= erc20Contract.balanceOf(buyerAddress), "NftMarket._settle: is not sufficient balance");
        let erc20balanceOf = Mev.Exec(erc20Contract, "balanceOf", [buyerAddress])
        erc20balanceOf = BigInt(erc20balanceOf)
        require(amount <= erc20balanceOf, "NftMarket._settle: is not sufficient balance");
        // require(amount <= erc20Contract.allowance(buyerAddress, address(this)), "NftMarket._settle: is not allowanced");
        let erc20allowance = Mev.Exec(erc20Contract, "allowance", [buyerAddress, Mev.This()])
        erc20allowance = BigInt(erc20allowance)
        require(amount <= erc20allowance, "NftMarket._settle: is not allowanced");
        
        // erc20Contract.transferFrom(buyerAddress, address(_marketData), marketFee);
        // erc20Contract.transferFrom(buyerAddress, address(adminAddress), foundationRoyalty);
        // erc20Contract.transferFrom(buyerAddress, nftOwner, remainAmount);
        Mev.Exec(erc20Contract, "transferFrom", [buyerAddress, _marketData(), marketFee])
        Mev.Exec(erc20Contract, "transferFrom", [buyerAddress, adminAddress, foundationRoyalty])
        Mev.Exec(erc20Contract, "transferFrom", [buyerAddress, burnAddress, burnCap])
        Mev.Exec(erc20Contract, "transferFrom", [buyerAddress, nftOwner, remainAmount])
        
        // _marketData.transactCompleteItemInMarket(marketItem);
        Mev.Exec(_marketData(), "transactCompleteItemInMarket", [marketItem])
        // emit MarketItemCompleted(SaleType.BUY_NOW, address(marketItem.nft.tokenContract), marketItem.nft.tokenId, marketItem.seller, _msgSender(), address(erc20Contract), amount);
    }
}

function _calcBurnFee(amount) {
    amount = BigInt(amount)
    return (amount * _burnFee()) / MAX_PERCENT;
}

function _calcRoyaltyFee(amount) {
    amount = BigInt(amount)
    return (amount * _royaltyFee()) / MAX_PERCENT;
}

function _calcMarketFee(amount) {
    amount = BigInt(amount)
    return (amount * _marketFee()) / MAX_PERCENT;
}
