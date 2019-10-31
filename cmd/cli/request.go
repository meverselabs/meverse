package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	uuid "github.com/satori/go.uuid"

	"github.com/fletaio/fleta/service/apiserver"
)

func DoRequest(hostURL string, Method string, Params []interface{}) (interface{}, error) {
	id := uuid.NewV1().String()
	req := &apiserver.JRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  Method,
		Params:  Params,
	}
	bs, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	r, err := http.Post(hostURL+"/api/endpoints/http", "application/json", bytes.NewReader(bs))
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	var res apiserver.JRPCResponse
	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		return nil, err
	}
	if res.Error != nil {
		return nil, errors.New(fmt.Sprint(res.Error))
	} else {
		return res.Result, nil
	}
}
