const tagOwner  = "01"
const tagABI    = "04"

//////////////////////////////////////////////////
// Public Functions
//////////////////////////////////////////////////

function Init(owner) {
	owner = address(owner)
    require(owner !== address(0), "owner not provided")
	Mev.SetContractData(tagOwner, owner)
}

let k = [
    {
        "inputs": [
            {
                "internalType": "uint256",
                "name": "count",
                "type": "uint256"
            }
        ],
        "name": "mint",
        "outputs": [
            {
                "internalType": "uint256[]",
                "name": "",
                "type": "uint256[]"
            }
        ],
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "isOwner",
        "outputs": [
            {
                "internalType": "bool",
                "name": "",
                "type": "bool"
            }
        ],
        "payable": false,
        "stateMutability": "view",
        "type": "function"
    }
]

function onlyOwner() {
    let com = Mev.ContractData(tagOwner)
	let from = address(Mev.From())
    if (from != com) {
        throw `onlyOwner ${from} is not owner(${com})`
    }
}

function _contractKey(contract) {
    return tagABI+address(contract)
}
function setAbi(contract, abi)  {
	onlyOwner()
	contract = address(contract)
    try {
        JSON.parse(abi)
    } catch (error) {
        // not parseable abi
        throw error
    }
	Mev.SetContractData(_contractKey(contract), abi)
}

function abi(contract)  {
	return Mev.ContractData(_contractKey(contract))
}
