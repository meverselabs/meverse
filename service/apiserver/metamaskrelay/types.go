package metamaskrelay

// RPCTransaction represents a transaction that will serialize to the RPC representation of a transaction
type RPCTransaction struct {
	BlockHash        string      `json:"blockHash"`
	BlockNumber      string      `json:"blockNumber"`
	From             string      `json:"from"`
	Gas              string      `json:"gas"`
	GasPrice         string      `json:"gasPrice"`
	GasFeeCap        string      `json:"maxFeePerGas,omitempty"`
	GasTipCap        string      `json:"maxPriorityFeePerGas,omitempty"`
	Hash             string      `json:"hash"`
	Input            string      `json:"input"`
	Nonce            interface{} `json:"nonce"`
	To               interface{} `json:"to"`
	TransactionIndex string      `json:"transactionIndex"`
	Value            string      `json:"value"`
	Type             string      `json:"type,omitempty"`
	Accesses         string      `json:"accessList,omitempty"`
	ChainID          string      `json:"chainId,omitempty"`
	V                string      `json:"v"`
	R                string      `json:"r"`
	S                string      `json:"s"`
}
