package txsearch

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/extern/txparser"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (t *TxSearch) BlockHeight(bh hash.Hash256) (uint32, error) {
	v, err := t.db.Get(append([]byte{tagBlockHash}, bh[:]...), nil)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	if len(v) == 0 {
		return 0, errors.New("not exist block")
	}

	if len(v) == 4 {
		return bin.Uint32(v), nil
	}
	return 0, errors.New("invalid length")
}

func (t *TxSearch) Block(i uint32) (*types.Block, error) {
	return t.st.Block(i)
}

func (t *TxSearch) TxIndex(th hash.Hash256) (TxID, error) {
	v, err := t.db.Get(append([]byte{tagTxHash}, th[:]...), nil)
	if err != nil && errors.Cause(err) != leveldb.ErrNotFound {
		return TxID{0, 0, nil}, errors.WithStack(err)
	}

	if len(v) == 0 {
		v, err = t.db.Get(toTxFailKey(th), nil)
		if err != nil {
			return TxID{0, 0, nil}, errors.WithStack(err)
		} else if len(v) > 0 {
			is, err := bin.TypeReadAll(v, 2)
			if err != nil {
				return TxID{0, 0, nil}, err
			}
			height := is[0].(uint32)
			errstr := is[1].(string)
			return TxID{height, 0, errors.New(errstr)}, ErrFailTx //fail tx
		}
		return TxID{0, 0, nil}, errors.New("not exist tx")
	}

	if len(v) == 6 {
		return TxID{bin.Uint32(v[:4]), bin.Uint16(v[4:]), nil}, nil
	}
	if len(v) == 0 {
		return TxID{0, 0, nil}, errors.New("invalid length")
	}
	return TxID{0, 0, nil}, nil
}

func (t *TxSearch) Reward(cont, rewarder common.Address) (*amount.Amount, error) {
	key := make([]byte, 41)
	key[0] = tagEventReward
	copy(key[1:], cont[:])
	copy(key[21:], rewarder[:])
	bs, err := t.db.Get(key, nil)
	if err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}
	am := amount.NewAmountFromBytes(bs)
	return am, nil
}

func (t *TxSearch) DailyReward(cont, rewarder common.Address, index int) (map[string]*amount.Amount, error) {
	key := make([]byte, 41)
	key[0] = tagDailyReward
	copy(key[1:], cont[:])
	copy(key[21:], rewarder[:])

	startTm := time.Date(1900, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	du := time.Since(startTm)
	dayDiff := du / time.Hour / 24
	if dayDiff < 0 {
		return nil, errors.New("invalid block Timestamp")
	}
	days := uint32(dayDiff) + 1

	from := make([]byte, 45)
	to := make([]byte, 45)
	copy(from, key)
	copy(to, key)

	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, days-uint32((index-1)*20))
	copy(to[41:], bs)

	binary.BigEndian.PutUint32(bs, days-uint32(index*20))
	copy(from[41:], bs)

	ams := map[string]*amount.Amount{}
	iter := t.db.NewIterator(&util.Range{Start: from, Limit: to}, nil)
	for iter.Next() {
		bs := iter.Value()
		am := amount.NewAmountFromBytes(bs)
		kbs := iter.Key()

		diff := binary.BigEndian.Uint32(kbs[41:])
		tTm := time.Date(1900, time.Month(1), int(diff)+1, 0, 0, 0, 0, time.UTC)
		ams[tTm.Format("2006-01-02")] = am
	}
	return ams, nil
}

// func (t *TxSearch) Holder(token common.Address) (map[common.Address]*amount.Amount, error) {
// 	ctx := t.cn.NewContext()
// 	tagTokenAmount := byte(0x10)
// 	bs := ctx.Data(token, common.Address{}, []byte{tagTokenAmount})
// 	return nil, nil
// }

func (t *TxSearch) Tx(height uint32, index uint16) (map[string]interface{}, error) {
	b, err := t.Block(height)
	if err != nil {
		return nil, err
	}
	if len(b.Body.Transactions) <= int(index) {
		return nil, errors.New("tx not found")
	}

	tx := b.Body.Transactions[int(index)]

	m := map[string]interface{}{
		"Hash":        tx.Hash(height).String(),
		"Height":      height,
		"Index":       index,
		"ChainID":     tx.ChainID,
		"Timestamp":   tx.Timestamp,
		"Seq":         tx.Seq,
		"To":          tx.To,
		"Method":      tx.Method,
		"GasPrice":    tx.GasPrice,
		"UseSeq":      tx.UseSeq,
		"IsEtherType": tx.IsEtherType,
		"From":        tx.From,
	}
	if tx.IsEtherType {
		ctx := t.cn.NewContext()
		data, _, err2 := types.TxArg(ctx, tx)
		if err2 == nil {
			m["Args"] = data
		}
	} else {
		m["Args"], err = bin.TypeReadAll(tx.Args, -1)
		if err != nil {
			m["Args"] = []interface{}{tx.Args}
		}
	}

	// TXID := types.TransactionIDBytes(height, index)

	// v, _ := t.db.Get(append([]byte{tagEvent}, TXID...), nil)
	if len(b.Body.Events) > 0 {
		find := index

		start := 0
		pin := start
		end := len(b.Body.Events)

		for len(b.Body.Events) > 0 && b.Body.Events[pin].Index != find {
			nextPin := (start + end) / 2
			if nextPin == pin {
				nextPin++
			}
			if nextPin >= end || start == end {
				break
			}
			nIndex := b.Body.Events[nextPin].Index
			if nIndex > find {
				end = nextPin
			} else if nIndex < find {
				start = nextPin
			}
			pin = nextPin
		}

		if b.Body.Events[pin].Index != find {
			m["ResultErr"] = errors.Errorf("invalid Event Index want %v, but not exist", index)
			return m, nil
		}

		for b.Body.Events[pin].Type != types.EventTagTxMsg && b.Body.Events[pin].Index == find && pin > 0 && b.Body.Events[pin-1].Index == find {
			pin--
		}

		if b.Body.Events[pin].Type == types.EventTagTxMsg && b.Body.Events[pin].Index == find {
			en := b.Body.Events[index]
			m["Result"], err = bin.TypeReadAll(en.Result, -1)
			if err != nil {
				m["ResultErr"] = err
			}
			pin++
		}

		event := []*MethodCallEvent{}
		for i := pin; i < len(b.Body.Events) && b.Body.Events[i].Index == find; i++ {
			bf := bytes.NewBuffer(b.Body.Events[i].Result)
			mc := &MethodCallEvent{}
			if _, err := mc.ReadFrom(bf); err == nil {
				cont, err := t.cn.NewContext().Contract(mc.To)
				if err == nil {
					if cntName, ok := cont.(ContractName); ok {
						mc.ToName = cntName.Name()
					} else if cntName, ok := cont.(ContractNameCC); ok {
						ctx := t.cn.NewContext()
						cc := ctx.ContractContext(cont, common.Address{})
						mc.ToName = cntName.Name(cc)
					}
				}
				event = append(event, mc)
			}
		}
		if len(event) > 0 {
			m["Event"] = event
		}
	}

	return m, nil
}

type BlockInfo struct {
	Height    uint32
	Hash      string
	TxLen     uint16
	Timestamp uint64
}

func (t *TxSearch) BlockList(index int) []*BlockInfo {
	h := t.st.Height()
	start := h
	if h < uint32(index*20) {
		start = 0
	} else {
		start = h - uint32(index*20)
	}

	li := []*BlockInfo{}
	for i := start; i > start-20; i-- {
		b, err := t.Block(i)
		if err == nil {
			li = append(li, &BlockInfo{
				Height:    i,
				Hash:      b.Header.ContextHash.String(),
				TxLen:     uint16(len(b.Body.Transactions)),
				Timestamp: b.Header.Timestamp,
			})
		}
	}

	return li
}

func (t *TxSearch) TxSize() uint64 {
	n := common.Address{}
	var aik addrIndexKey
	aik[0] = tagID
	copy(aik[1:], n[:])

	bs, _ := t.db.Get(aik[:], nil)
	if len(bs) != 8 {
		bs = make([]byte, 8)
	}
	return bin.Uint64(bs)
}

func (t *TxSearch) TxList(index int) ([]TxList, error) {
	n := common.Address{}
	tlen, from, to := t.getRange(tagID, n[:], index)
	txs := make([]TxList, tlen)
	iter := t.db.NewIterator(&util.Range{Start: from, Limit: to}, nil)
	var i int
	for iter.Next() {
		i++
		// key := iter.Key()
		bs := iter.Value()
		data, err := bin.TypeReadAll(bs, 4)
		if err != nil {
			continue
		}
		txid := data[0].([]byte)
		from := data[1].(common.Address)
		to := data[2].(common.Address)
		method := data[3].(string)
		h, index, err := types.ParseTransactionIDBytes(txid)
		if err != nil {
			continue
		}
		str := types.TransactionID(h, index)
		txs[int(tlen)-i] = TxList{
			TxID:   str,
			From:   from.String(),
			To:     to.String(),
			Method: method,
		}
	}
	return txs, nil
}

func (t *TxSearch) AddressTxList(From common.Address, index int) ([]TxList, error) {
	tlen, from, to := t.getRange(tagAddress, From[:], index)
	txs := make([]TxList, tlen)

	iter := t.db.NewIterator(&util.Range{Start: from, Limit: to}, nil)
	var i int
	for iter.Next() {
		i++
		data, method, h, index, err := t.txDateSet(iter.Value(), 2)
		if err != nil {
			continue
		}
		id := int(tlen) - i
		txs[id] = TxList{
			TxID:   types.TransactionID(h, index),
			Method: method,
			From:   From.String(),
		}
		if len(data) == 2 {
			continue
		}
		token, ok := data[2].(common.Address)
		if !ok {
			continue
		}
		txs[id].Contract = token.String()
		switch method {
		case "Transfer", "TransferFrom":
			FromType := data[3].(uint8)
			To := data[4].(common.Address)
			am := data[5].(*amount.Amount)
			txs[id].Amount = am.String()
			if FromType == 0 {
				txs[id].From = From.String()
				txs[id].To = To.String()
			} else {
				txs[id].To = From.String()
				txs[id].From = To.String()
			}
		case "Approve":
			To := data[2].(common.Address)
			txs[id].To = To.String()
		case "CreateAlpha", "CreateSigma", "CreateOmega":
		case "Revoke":
		case "Stake":
			txs[id].To = data[2].(common.Address).String()
		case "Unstake":
			txs[id].To = data[2].(common.Address).String()
		}
	}
	return txs, nil
}

// func (t *TxSearch) TouchContract(cont common.Address, addr common.Address) error {
// 	contKey := addrKey(tagContractUsers, cont[:])
// 	bs, err := t.db.Get(contKey, nil)
// 	if err != nil {
// 		return err
// 	}
// 	if len(bs) != 8 {
// 		bs = make([]byte, 8)
// 	}

// }

// func (t *TxSearch) TokenTouchedAddrs(cont common.Address) ([]common.Address, error) {
// 	contKey := addrKey(tagContractUsers, cont[:])
// 	bs, _ := t.db.Get(contKey, nil)
// 	if len(bs) != 8 {
// 		initTokenTouchedAddrs(cont)
// 		bs = make([]byte, 8)
// 	}
// 	userCount := bin.Uint64(bs)

// 	t.db.Get()
// 	tlen, from, to := t.getRange(tagDefault, cont[:], index)
// 	txs := make([]TxList, tlen)

// 	iter := t.db.NewIterator(&util.Range{Start: from, Limit: to}, nil)
// 	var i int
// 	for iter.Next() {
// 		i++
// 		data, method, h, index, err := t.txDateSet(iter.Value(), 3)
// 		if err != nil {
// 			continue
// 		}
// 		id := int(tlen) - i
// 		txs[id] = TxList{
// 			TxID:   types.TransactionID(h, index),
// 			Method: method,
// 			From:   From.String(),
// 		}
// 		to, ok := data[2].(common.Address)
// 		if !ok {
// 			continue
// 		}
// 		txs[id].To = to.String()
// 	}
// 	return txs, nil
// }

func (t *TxSearch) TokenTxList(From common.Address, index int) ([]TxList, error) {
	tlen, from, to := t.getRange(tagDefault, From[:], index)
	txs := make([]TxList, tlen)

	iter := t.db.NewIterator(&util.Range{Start: from, Limit: to}, nil)
	var i int
	for iter.Next() {
		i++
		data, method, h, index, err := t.txDateSet(iter.Value(), 3)
		if err != nil {
			continue
		}
		id := int(tlen) - i
		txs[id] = TxList{
			TxID:   types.TransactionID(h, index),
			Method: method,
			From:   From.String(),
		}
		to, ok := data[2].(common.Address)
		if !ok {
			continue
		}
		txs[id].To = to.String()
	}
	return txs, nil
}

func (*TxSearch) txDateSet(bs []byte, paramCount int) ([]interface{}, string, uint32, uint16, error) {
	data, err := bin.TypeReadAll(bs, paramCount)
	if err != nil {
		return nil, "", 0, 0, err
	}
	method, ok := data[0].(string)
	if !ok {
		return nil, "", 0, 0, errors.New("not exist method")
	}
	TxID, ok := data[1].([]byte)
	if !ok {
		return nil, "", 0, 0, errors.New("not exist TxID")
	}
	h, index, err := types.ParseTransactionIDBytes(TxID)
	if err != nil {
		return nil, "", 0, 0, err
	}
	return data, method, h, index, nil
}

func (t *TxSearch) TransferTxList(token, From common.Address, index int) ([]TxList, error) {
	tlen, fromKey, toKey := t.getRange41(tagTransfer, append(token[:], From[:]...), index)

	iter := t.db.NewIterator(&util.Range{Start: fromKey, Limit: toKey}, nil)

	txs := make([]TxList, tlen)
	var i int
	for iter.Next() {
		i++
		bs := iter.Value()
		data, err := bin.TypeReadAll(bs, -1)
		if err != nil {
			continue
		}
		method := data[0].(string)
		TxID := data[1].([]byte)
		h, index, err := types.ParseTransactionIDBytes(TxID)
		if err != nil {
			continue
		}
		TxIDStr := types.TransactionID(h, index)
		txs[int(tlen)-i] = TxList{
			TxID:     TxIDStr,
			Method:   method,
			Contract: token.String(),
		}
		switch method {
		case "Transfer", "TransferFrom", "TokenIn", "TokenIndexIn", "TokenOut":
			FromType := data[2].(uint8)
			To := data[3].(common.Address)
			am := data[4].(*amount.Amount)
			txs[int(tlen)-i].Amount = am.String()
			if FromType == 0 {
				txs[int(tlen)-i].From = From.String()
				txs[int(tlen)-i].To = To.String()
			} else {
				txs[int(tlen)-i].To = From.String()
				txs[int(tlen)-i].From = To.String()
			}
		}
	}
	return txs, nil
}

func (t *TxSearch) getRange(b byte, From []byte, index int) (uint64, []byte, []byte) {
	var aik addrIndexKey
	aik[0] = b
	copy(aik[1:], From[:])
	return t._getRange(aik[:], From, index)
}

func (t *TxSearch) getRange41(b byte, From []byte, index int) (uint64, []byte, []byte) {
	var aik addr41IndexKey
	aik[0] = b
	copy(aik[1:], From[:])
	return t._getRange(aik[:], From, index)
}

func (t *TxSearch) _getRange(aik []byte, From []byte, index int) (uint64, []byte, []byte) {
	bs, _ := t.db.Get(aik, nil)
	if len(bs) != 8 {
		bs = make([]byte, 8)
	}
	s := bin.Uint64(bs)
	_to := int64(s) - int64(index*20) + 1
	if _to < 1 {
		_to = 1
	}
	_from := int64(s) - int64(index*20) - 20 + 1
	if _from < 1 {
		_from = 1
	}
	from := uint64(_from)
	to := uint64(_to)
	tlen := to - from
	bs = make([]byte, 8)
	binary.BigEndian.PutUint64(bs, from)
	fromKey := append(aik, bs...)
	binary.BigEndian.PutUint64(bs, to)
	toKey := append(aik, bs...)
	return tlen, fromKey, toKey
}

func (t *TxSearch) TokenOutList(height uint32) (interface{}, error) {
	max, err := hex.DecodeString("ffffffff")
	if err != nil {
		return nil, err
	}
	max = append([]byte{tagTokenOut}, max...)

	hbs := make([]byte, 4)
	binary.BigEndian.PutUint32(hbs, height)
	min := append([]byte{tagTokenOut}, hbs...)

	iter := t.db.NewIterator(&util.Range{Start: min, Limit: max}, nil)

	txs := []map[string]string{}
	for iter.Next() {
		kbs := iter.Key()
		t := map[string]string{}

		height := binary.BigEndian.Uint32(kbs[1:])
		t["Height"] = fmt.Sprintf("%v", height)

		is, err := bin.TypeReadAll(iter.Value(), -1)
		if err != nil {
			continue
		}
		TxID := is[0].([]byte)
		from := is[1].(common.Address)
		Platform := is[2].(string)
		withdrawAddress := is[3].(common.Address)
		Deposit, ok4 := is[4].(*big.Int)

		t["TxID"] = hex.EncodeToString(TxID)
		t["From"] = from.String()
		t["Platform"] = Platform
		t["To"] = withdrawAddress.String()
		if ok4 {
			t["DepositHex"] = hex.EncodeToString(Deposit.Bytes())
		} else {
			t["DepositHex"] = ""
		}
		txs = append(txs, t)
	}

	return txs, nil
}

func (t *TxSearch) TokenLeaveList(height uint32) (interface{}, error) {
	max, err := hex.DecodeString("ffffffff")
	if err != nil {
		return nil, err
	}
	max = append([]byte{tagTokenLeave}, max...)

	hbs := make([]byte, 4)
	binary.BigEndian.PutUint32(hbs, height)
	min := append([]byte{tagTokenLeave}, hbs...)

	iter := t.db.NewIterator(&util.Range{Start: min, Limit: max}, nil)

	txs := []map[string]string{}
	for iter.Next() {
		kbs := iter.Key()
		t := map[string]string{}

		height := binary.BigEndian.Uint32(kbs[1:])
		t["Height"] = fmt.Sprintf("%v", height)

		is, err := bin.TypeReadAll(iter.Value(), -1)
		if err != nil {
			continue
		}
		TxID, ok := is[0].([]byte)
		if !ok {
			continue
		}
		CoinTXID, ok := is[1].(string)
		if !ok {
			continue
		}
		ERC20TXID, ok := is[2].(string)
		if !ok {
			continue
		}
		Platform, ok := is[3].(string)
		if !ok {
			continue
		}

		t["TxID"] = hex.EncodeToString(TxID)
		t["CoinTXID"] = CoinTXID
		t["ERC20TXID"] = ERC20TXID
		t["Platform"] = Platform
		txs = append(txs, t)
	}

	return txs, nil
}

func (t *TxSearch) BridgeTxList(contStr common.Address, height uint32, to string) (interface{}, error) {
	maxHeight, err := hex.DecodeString("ffffffff")
	if err != nil {
		return nil, err
	}
	if to != "latest" {
		u64, err := strconv.ParseUint(to, 10, 32)
		if err != nil {
			return nil, err
		}
		v := uint32(u64)
		hbs := make([]byte, 4)
		binary.BigEndian.PutUint32(hbs, v)
		maxHeight = hbs[:]
	}
	max := append([]byte{tagBridge}, contStr[:]...)
	max = append(max, maxHeight...)

	hbs := make([]byte, 4)
	binary.BigEndian.PutUint32(hbs, height)
	min := append([]byte{tagBridge}, contStr[:]...)
	min = append(min, hbs...)

	iter := t.db.NewIterator(&util.Range{Start: min, Limit: max}, nil)

	txs := []map[string]interface{}{}
	for iter.Next() {
		res := map[string]interface{}{}

		is, err := bin.TypeReadAll(iter.Value(), -1)
		if err != nil {
			continue
		}
		TxID, ok := is[0].([]byte)
		if !ok {
			continue
		}
		h, i, err := types.ParseTransactionIDBytes(TxID)
		if err != nil {
			continue
		}

		b, err := t.st.Block(h)
		if err != nil {
			return nil, err
		}
		if len(b.Body.Transactions) <= int(i) {
			continue
		}
		tx := b.Body.Transactions[int(i)]

		bHash := bin.MustWriterToHash(&b.Header)
		res["blockNumber"] = fmt.Sprintf("%v", b.Header.Height)
		res["timeStamp"] = fmt.Sprintf("%v", b.Header.Timestamp/1000)
		res["hash"] = tx.Hash(b.Header.Height).String()
		res["nonce"] = fmt.Sprintf("%v", tx.Seq)
		res["blockHash"] = bHash.String()
		res["transactionIndex"] = fmt.Sprintf("%v", i)
		res["from"] = tx.From.String()
		res["to"] = tx.To.String()
		res["value"] = "0"
		res["gas"] = "10000"
		res["gasPrice"] = tx.GasPrice.String()
		res["isError"] = "0"
		res["txreceipt_status"] = "1"
		res["contractAddress"] = ""
		res["cumulativeGasUsed"] = "100000000000000000"
		res["gasUsed"] = "10000"
		res["confirmations"] = fmt.Sprintf("%v", t.Height()-height)
		if tx.Method == "SendToGateway" {
			res["event"] = "SentToGateway"
		} else if tx.Method == "SendFromGateway" || tx.Method == "SendToRouterFromGateway" {
			res["event"] = "SentFromGateway"
		}

		var events []interface{}
		v, _ := t.db.Get(append([]byte{tagEvent}, TxID...), nil)
		if len(v) == 2 {
			index := bin.Uint16(v)
			if len(b.Body.Events) < int(index) {
			} else {
				en := b.Body.Events[index]
				events, _ = bin.TypeReadAll(en.Result, -1)
			}
		}

		var data []interface{}
		if tx.IsEtherType {
			etx, _, err := txparser.EthTxFromRLP(tx.Args)
			if err != nil {
				continue
			}
			res["input"] = hex.EncodeToString(etx.Data())
			etx, _, err = txparser.EthTxFromRLP(tx.Args)
			if err != nil {
				continue
			}
			if etx.Value().Cmp(amount.ZeroCoin.Int) > 0 && tx.To != *t.st.MainToken() {
				data = []interface{}{&amount.Amount{Int: etx.Value()}}
			} else {
				eData := etx.Data()
				if len(eData) > 0 {
					data, err = txparser.Inputs(eData)
					if err != nil {
						continue
					}
				}
			}
		} else {
			data, err = bin.TypeReadAll(tx.Args, -1)
			if err != nil {
				continue
			}
		}

		rv := t.makeReturnValue(tx, data, res, events)
		if rv != nil {
			res["returnValues"] = rv
		}
		txs = append(txs, res)
	}

	return txs, nil
}

func (*TxSearch) makeReturnValue(tx *types.Transaction, data []interface{}, res map[string]interface{}, events []interface{}) map[string]interface{} {
	returnValues := map[string]interface{}{}
	if tx.Method == "SendToGateway" {
		if len(data) != 5 {
			res["error"] = "SendToGateway not match params"
			return nil
		}

		if token, ok := data[0].(common.Address); ok {
			returnValues["_token"] = token
		}
		returnValues["_amount"] = getAmountOrBigIntString(data[1])
		if path, ok := data[2].([]common.Address); ok {
			returnValues["_path"] = path
		}
		returnValues["_summary"] = arrayToSliceHex(data[4])
		returnValues["_to"] = tx.To
		returnValues["_from"] = tx.From

	} else if (tx.Method == "SendFromGateway" || tx.Method == "SendToRouterFromGateway") && len(data) > 0 {
		if token, ok := data[0].(common.Address); ok {
			returnValues["_token"] = token
		}
		if to, ok := data[1].(common.Address); ok {
			returnValues["_to"] = to
		}
		// SendFromGateway         param: token 0, to 1, amt 2, path 3, fromChain 4, summary 5
		// SendToRouterFromGateway param: token 0, to 1, amountIn 2, amountOutMin 3, path 4, deadline 5, fromChain 6, summary 7
		routerCap := 0
		if tx.Method == "SendFromGateway" {
			amt := getAmountOrBigIntString(data[2])
			returnValues["_amountIn"] = amt
			returnValues["_amountOutMin"] = amt
		} else if tx.Method == "SendToRouterFromGateway" {
			returnValues["_amountIn"] = getAmountOrBigIntString(data[2])
			returnValues["_amountOutMin"] = getAmountOrBigIntString(data[3])
			routerCap = 1
		}
		if path, ok := data[3+routerCap].([]common.Address); ok {
			returnValues["_path"] = path
		}
		if tx.Method == "SendToRouterFromGateway" {
			routerCap = 2
		}
		returnValues["_summary"] = arrayToSliceHex(data[5+routerCap])

		returnValues["_from"] = tx.From
	}

	if len(events) == 1 {
		if seq, ok := events[0].(uint64); ok {
			returnValues["_sequence"] = seq
		}
	}
	return returnValues
}

func getAmountOrBigIntString(is interface{}) string {
	if amt, ok := is.(*amount.Amount); ok {
		return amt.Int.String()
	} else if amt, ok := is.(*big.Int); ok {
		return amt.String()
	}
	return ""
}

func arrayToSliceHex(is interface{}) string {
	param := reflect.ValueOf(is)
	bs := []byte{}
	if param.Kind() == reflect.Slice {
		bs = param.Bytes()
	} else if param.Type().Kind() == reflect.Array {
		l := param.Type().Len()
		for i := 0; i < l; i++ {
			val := param.Index(i).Interface()
			b := val.(byte)
			bs = append(bs, b)
		}
	}
	return hex.EncodeToString(bs)
}
