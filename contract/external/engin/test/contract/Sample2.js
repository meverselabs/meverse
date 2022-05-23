const Owner         = "01"
const Data          = "02"
const Data2         = "03"

function Init (owner) {
    Mev.Log(owner)
    if (typeof owner === "undefined") {
        throw "owner not provided"
    }
	Mev.SetContractData(Owner, owner)
}

function SetData(data, index) {
    let from = Mev.From()
    if (typeof index == "undefined") {
        Mev.SetAccountData(from, Data, data)
    } else if (index == "0") {
        Mev.SetAccountData(from, Data, data)
    } else if (index == "1") {
        Mev.SetAccountData(from, Data2, data)
    } else {
        throw "not supported index"
    }
}

function GetData(index) {
    let from = Mev.From()
    let msg
    if (typeof index == "undefined") {
        msg = Mev.AccountData(from, Data)
    } else if (index == "0") {
        msg = Mev.AccountData(from, Data)
    } else if (index == "1") {
        msg = Mev.AccountData(from, Data2)
    } else {
        throw "not supported index"
    }
    return msg
}
