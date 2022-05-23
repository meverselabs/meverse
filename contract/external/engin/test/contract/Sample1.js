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

function SetData(data) {
    let from = Mev.From()
    Mev.SetAccountData(from, Data, data)
}

function GetData(index) {
    let from = Mev.From()
    let msg
    msg = Mev.AccountData(from, Data)
    return msg
}
