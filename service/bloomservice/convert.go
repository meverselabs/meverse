package bloomservice

var convertMap map[string]map[string]string

func init() {
	convertMap = map[string]map[string]string{
		"token.TokenContract": {
			"Approve":      "Approval(address,address,uint256)",
			"Transfer":     "Transfer(address,address,uint256)",
			"TransferFrom": "Transfer(address,address,uint256)",
			"Mint":         "Transfer(address,address,uint256)",
			"Burn":         "Transfer(address,address,uint256)",
		},
	}

}
