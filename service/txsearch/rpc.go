package txsearch

import (
	"bytes"
	"reflect"
	"strings"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/service/apiserver"
	"github.com/pkg/errors"
)

func (t *TxSearch) SetupApi() error {
	s, err := t.api.JRPC("search")
	if err != nil {
		panic(err)
	}

	// chainID := "0xffff"
	s.Set("version", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return "v1.0.3", nil
	})
	s.Set("blocks", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		index, _ := arg.Int(0)
		return t.BlockList(index), nil
	})
	s.Set("block", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		index, err := arg.Uint32(0)
		if err != nil {
			h, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			h = strings.Replace(h, "0x", "", 1)
			index, err = t.BlockHeight(hash.HexToHash(h))
			if err != nil {
				return nil, err
			}
		}
		b, err := t.Block(index)
		if err != nil {
			return nil, err
		}
		m := map[string]interface{}{}
		m["Header"] = b.Header
		bodymap := map[string]interface{}{
			"Transactions":          b.Body.Transactions,
			"TransactionSignatures": b.Body.TransactionSignatures,
			"BlockSignatures":       b.Body.BlockSignatures,
		}
		Events := []interface{}{}
		for _, v := range b.Body.Events {
			m := map[string]interface{}{
				"Type":  v.Type.String(),
				"Index": v.Index,
			}
			bm := map[common.Address][]byte{}
			switch v.Type {
			case types.EventTagCallHistory:
				bf := bytes.NewBuffer(v.Result)
				mc := &types.MethodCallEvent{}
				if _, err := mc.ReadFrom(bf); err != nil {
					m["Error"] = err
					continue
				} else {
					m["callHistory"] = mc
					Events = append(Events, m)
				}
			case types.EventTagReward:
				err := types.UnmarshalAddressBytesMap(v.Result, bm)
				if err != nil {
					m["Error"] = err
					Events = append(Events, m)
					continue
				}
				rm := map[common.Address]interface{}{}
				for addr, r := range bm {
					rem := map[common.Address]*amount.Amount{}
					err := types.UnmarshalAddressAmountMap(r, rem)
					if err != nil {
						rm[addr] = err
					} else {
						rm[addr] = rem
					}
				}
				m["Reward"] = rm
				Events = append(Events, m)
			case types.EventTagTxMsg:
				ins, err := bin.TypeReadAll(v.Result, 1)
				if err != nil {
					m["Error"] = err
					Events = append(Events, m)
					continue
				}
				m["message"] = ins
				Events = append(Events, m)
			}
		}
		bodymap["Events"] = Events
		m["Body"] = bodymap
		return m, nil
	})
	s.Set("txSize", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return t.TxSize(), nil
	})
	s.Set("txs", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		index, _ := arg.Int(0)
		return t.TxList(index)
	})
	s.Set("tx", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		height, err := arg.Uint32(0)
		index, _ := arg.Uint16(1)
		if err != nil {
			h, err := arg.String(0)
			if err != nil {
				return nil, err
			}
			if len(h) == 12 {
				height, index, err = types.ParseTransactionID(h)
				if err != nil {
					return nil, err
				}
			} else {
				h = strings.Replace(h, "0x", "", 1)
				txID, err := t.TxIndex(hash.HexToHash(h))
				if err != nil {
					if errors.Cause(err) != txID.Err {
						return nil, txID.Err
					}
					return nil, err
				}
				height = txID.Height
				index = txID.Index
			}
		}
		return t.Tx(height, index)
	})
	s.Set("addressSize", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return t.TagSize(tagAddress, arg)
	})
	s.Set("address", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		addrStr, err := arg.String(0)
		if err != nil {
			return nil, err
		}
		index, err := arg.Int(1)
		if err != nil {
			return nil, err
		}
		addr := common.HexToAddress(addrStr)
		return t.AddressTxList(addr, index)
	})
	s.Set("tokenSize", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return t.TagSize(tagDefault, arg)
	})
	s.Set("token", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		addrStr, err := arg.String(0)
		if err != nil {
			return nil, err
		}
		index, err := arg.Int(1)
		if err != nil {
			return nil, err
		}
		addr := common.HexToAddress(addrStr)
		return t.TokenTxList(addr, index)
	})
	s.Set("transferList", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		tokenStr, err := arg.String(0)
		if err != nil {
			return nil, err
		}
		addrStr, err := arg.String(1)
		if err != nil {
			return nil, err
		}
		index, err := arg.Int(2)
		if err != nil {
			return nil, err
		}
		addr := common.HexToAddress(addrStr)
		token := common.HexToAddress(tokenStr)
		return t.TransferTxList(token, addr, index)
	})
	s.Set("reward", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		tokenStr, err := arg.String(0)
		if err != nil {
			return nil, err
		}
		addrStr, err := arg.String(1)
		if err != nil {
			return nil, err
		}
		token := common.HexToAddress(tokenStr)
		addr := common.HexToAddress(addrStr)
		return t.Reward(token, addr)
	})
	s.Set("dailyReward", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		tokenStr, err := arg.String(0)
		if err != nil {
			return nil, err
		}
		addrStr, err := arg.String(1)
		if err != nil {
			return nil, err
		}
		index, err := arg.Int(2)
		if err != nil {
			return nil, err
		}
		token := common.HexToAddress(tokenStr)
		addr := common.HexToAddress(addrStr)
		return t.DailyReward(token, addr, index)
	})
	s.Set("tokenOuts", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		height, err := arg.Uint32(0)
		if err != nil {
			return nil, err
		}
		return t.TokenOutList(height)
	})
	s.Set("tokenLeaves", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		height, err := arg.Uint32(0)
		if err != nil {
			return nil, err
		}
		return t.TokenLeaveList(height)
	})
	s.Set("bridgeTxs", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		contStr, err := arg.String(0)
		if err != nil {
			return nil, err
		}
		From, err := arg.Uint32(1)
		if err != nil {
			return nil, err
		}
		To, err := arg.String(2)
		if err != nil {
			return nil, err
		}
		cont := common.HexToAddress(contStr)
		return t.BridgeTxList(cont, From, To)
	})
	s.Set("contracts", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		var contracts = make(map[common.Address]string)
		cs, err := t.st.Contracts()
		if err != nil {
			return nil, err
		}
		for _, c := range cs {
			rt := reflect.TypeOf(c)
			sliceName := strings.Split(rt.String(), ".")
			contracts[c.Address()] = strings.Replace(sliceName[0], "*", "", -1)
		}
		return contracts, nil
	})

	return nil
}

func (t *TxSearch) TagSize(tag byte, arg *apiserver.Argument) (uint64, error) {
	addrStr, err := arg.String(0)
	if err != nil {
		return 0, err
	}
	addr := common.HexToAddress(addrStr)

	var aik addrIndexKey
	aik[0] = tag
	copy(aik[1:], addr[:])

	bs, _ := t.db.Get(aik[:], nil)
	if len(bs) != 8 {
		bs = make([]byte, 8)
	}
	return bin.Uint64(bs), nil
}
