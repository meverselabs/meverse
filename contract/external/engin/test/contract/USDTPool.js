
const Owner         = "01"
const DepositToken  = "02"
const DepositAmount = "03"
const DepositLock   = "04"
const WithdrawLock  = "05"
const AccountLength = "06"
const AccountIndex  = "07"
const HolderList    = "08"

function Init (owner, token, amt) {
    Log(owner, token, amt)
    if (typeof owner === "undefined") {
        throw "owner not provided"
    }
    if (typeof token === "undefined") {
        throw "token not provided"
    }
    if (typeof amt === "undefined") {
        throw "amt not provided"
    }
	SetContractData(Owner, owner)
	SetContractData(DepositToken, token)
	SetContractData(DepositAmount, amt)
    SetContractData(DepositLock, "false")
    SetContractData(WithdrawLock, "true")
}

function Deposit() {
    if (ContractData(DepositLock) == "true") {
        throw "Locked"
    }
    let token = ContractData(DepositToken)
    let amt = ContractData(DepositAmount)
    let from = From()
	let deposit = AccountData(from, DepositToken)
    if (deposit) {
        throw "aleady deposit"
    }
    let _this = This()
    addDepositUser(from)
    let msg = SetContractData(HolderList+from, "1")
    msg = SetAccountData(from, DepositToken, token)
    deposit = AccountData(from, DepositToken)
    return Exec(token, "TransferFrom", [from, _this, amt])
}

function Holder() {
    return ContractData(AccountLength)
}

function Holders() {
    let len = ContractData(AccountLength)
    len = len-0
    let data = []
    for (var i = 0 ; i < len ; i++) {
        let addr = ContractData(AccountIndex+i)
        data.push(addr)
    }
    return data
}

function HolderMap() {
    let len = ContractData(AccountLength)
    len = len-0
    let amt = ContractData(DepositAmount)
    let data = {}
    for (var i = 0 ; i < len ; i++) {
        let addr = ContractData(AccountIndex+i)
        data[addr] = amt
    }
    return data
}

function IsHolder(addr) {
    return true
    return ContractData(HolderList+addr) == "1"
}

function GetCounter() {
    return 0
}

function addDepositUser(from) {
    let count = ContractData(AccountLength)
    if (count == "") {
        count = 0
    }
    SetContractData(AccountLength, (count-0) + 1)
    SetContractData(AccountIndex+count, from)
}

function GetDepositUser(index) {
    return ContractData(AccountIndex+index)
}

function LockDeposit() {
    if (From() != ContractData(Owner)) {
        return "Not Owner"
    }
    SetContractData(DepositLock, "true")
}

function UnlockWithdraw() {
    if (From() != ContractData(Owner)) {
        return "Not Owner"
    }
    SetContractData(WithdrawLock, "false")
}

function ReclaimToken(token, amt) {
    let from = From()
    if (from != ContractData(Owner)) {
        throw "Not Owner"
    }
    return Exec(token, "Transfer", [from, amt])
}

function Withdraw() {
    if (ContractData(WithdrawLock) == "true") {
        throw "Locked"
    }
    let token = ContractData(DepositToken)
    let amt = ContractData(DepositAmount)
	let deposit = AccountData(From(), DepositToken)
    if (!deposit) {
        throw "No Deposit"
    }
    return Exec(token, "Transfer", [From(), amt])
}
