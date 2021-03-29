package apiserver

import (
	"encoding/json"
	"sync"
)

// Handler handles a rpc method
type Handler func(ID interface{}, arg *Argument) (interface{}, error)

// JRPCSub provides the json rpc feature of the sub name
type JRPCSub struct {
	sync.Mutex
	funcMap map[string]Handler
}

// NewJRPCSub returns a JRPCSub
func NewJRPCSub() *JRPCSub {
	s := &JRPCSub{
		funcMap: map[string]Handler{},
	}
	return s
}

// Set sets a handler of the method
func (s *JRPCSub) Set(Method string, h Handler) {
	s.funcMap[Method] = h
}

// JRPCRequest is a jrpc request
type JRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// jRPCRequest is a jrpc request
type jRPCRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      interface{}    `json:"id"`
	Method  string         `json:"method"`
	Params  []*json.Number `json:"params"`
}

// jRPCRequest2 is a jrpc request
type jRPCRequest2 struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// JRPCResponse is a jrpc response
type JRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error"`
}
