package testlib

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/meverselabs/meverse/service/apiserver"
)

// JsonClient is a json-rpc senter
type JsonClient struct {
	tb *TestBlockChain
}

// NewJsonClient makes a new JsonClient
func NewJsonClient(tb *TestBlockChain) *JsonClient {
	return &JsonClient{tb: tb}
}

// GetBlockByNumber executes an eth_getBlockByNumber json-rpc
func (jc *JsonClient) GetBlockByNumber(height uint32, isFull bool) map[string]interface{} {
	req := &apiserver.JRPCRequest{
		JSONRPC: "2.0",
		ID:      "100",
		Method:  "eth_getBlockByNumber",
		Params:  []interface{}{height, isFull},
	}
	return jc.tb.HandleJRPC(req).(*apiserver.JRPCResponse).Result.(map[string]interface{})
}

// GetTransactionReceipt executes an eth_getTransactionReceipt json-rpc
func (jc *JsonClient) GetTransactionReceipt(hash common.Hash) map[string]interface{} {
	req := &apiserver.JRPCRequest{
		JSONRPC: "2.0",
		ID:      "101",
		Method:  "eth_getTransactionReceipt",
		Params:  []interface{}{hash.Hex()},
	}
	return jc.tb.HandleJRPC(req).(*apiserver.JRPCResponse).Result.(map[string]interface{})
}

// GetLogs executes an eth_getLogs json-rpc
// filterMap := map[string]interface{}{}
// filterMap["address"] = address.String()
// filterMap["blockHash"] = blockHash.String()
// filterMap["fromBlock"] = fmt.Sprintf("0x%x", *big.Int)
// filterMap["toBlock"] = fmt.Sprintf("0x%x", *big.Int)
// filterMap["topics"] = []interface{}{[]interface{}{transferHash.String()}} or
// 						 []interface{}{transferHash.String()}
// topics :  service/bloomservice/filter.go FilterQuery struct 참조
func (jc *JsonClient) GetLogs(filterMap map[string]interface{}) []*types.Log {

	req := &apiserver.JRPCRequest{
		JSONRPC: "2.0",
		ID:      "102",
		Method:  "eth_getLogs",
		Params:  []interface{}{filterMap},
	}
	return jc.tb.HandleJRPC(req).(*apiserver.JRPCResponse).Result.([]*types.Log)
}

// ViewCall executes an view.call json-rpc to non-evm contract
func (jc *JsonClient) ViewCall(to *common.Address, method string, params ...any) interface{} {
	req := &apiserver.JRPCRequest{
		JSONRPC: "2.0",
		ID:      "103",
		Method:  "view.call",
		Params:  []interface{}{*to, method, params},
	}
	return jc.tb.HandleJRPC(req).(*apiserver.JRPCResponse).Result
}
