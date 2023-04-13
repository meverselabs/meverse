package txsearch

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	etypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/ctypes"

	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/extern/txparser"
	"github.com/syndtr/goleveldb/leveldb"
)

type readB struct {
	Hkey  byte
	key   []byte
	value []byte
	tag   string
}

var txLen int

type addrIndexKey [21]byte
type addr41IndexKey [41]byte

func (t *TxSearch) ReadBlock(b *types.Block) (err error) {
	batch := new(leveldb.Batch)
	defer func() {
		if err == nil {
			if b.Header.Height%10000 == 0 {
				plog("setHeight", b.Header.Height, "txlen", txLen)
				txLen = 0
			}
			t.setHeight(b.Header.Height)
		}
		err := t.db.Write(batch, nil)
		if err != nil {
			panic(err)
		}
	}()

	if t.Height()+1 != b.Header.Height {
		return &ErrIsNotNextBlock{}
	}

	heightbs := bin.Uint32Bytes(b.Header.Height)
	l := []*readB{}
	bHash := bin.MustWriterToHash(&b.Header)
	l = append(l, &readB{tagBlockHash, bHash[:], heightbs, fmt.Sprintln("block", b.Header.Height)})
	txLen += len(b.Body.Transactions)

	indexMap := map[addrIndexKey]uint64{}
	index41Map := map[addr41IndexKey]uint64{}
	for i, tx := range b.Body.Transactions {
		TXID := types.TransactionIDBytes(b.Header.Height, uint16(i))
		TxTag := fmt.Sprintln("tx height:", b.Header.Height, "index:", i)
		l = append(l, &readB{tagTxHash, tx.Hash(b.Header.Height).Bytes(), TXID, TxTag})

		if tx.VmType == types.Evm {
			etx := new(etypes.Transaction)
			if err := etx.UnmarshalBinary(tx.Args); err != nil {
				return err
			}

			dataBs := etx.Data()
			if len(dataBs) == 0 {
				tx.Method = "Transfer"
			} else {
				m := txparser.Abi(hex.EncodeToString(dataBs[:4]))
				if m.Name != "" {
					tx.Method = strings.ToUpper(string(m.Name[0])) + m.Name[1:]
				}
			}

			l = append(l, &readB{tagTxHash, etx.Hash().Bytes(), TXID, TxTag})
		} else {
			l = append(l, &readB{tagTxHash, tx.Hash(b.Header.Height).Bytes(), TXID, TxTag})
		}
		t.saveTx(indexMap, index41Map, batch, tx, b.Header.Height, TXID)
	}
	for k, v := range indexMap {
		batch.Put(k[:], bin.Uint64Bytes(v))
	}
	for k, v := range index41Map {
		batch.Put(k[:], bin.Uint64Bytes(v))
	}

	if b.Header.Height%20 == 0 {
		days := makeBlockDay(b)
		for _, en := range b.Body.Events {
			if en.Type == ctypes.EventTagReward {
				t.saveRewardEvent(en, batch, days)
			}
		}
	}

	for _, rb := range l {
		batch.Put(append([]byte{rb.Hkey}, rb.key...), rb.value)
	}

	return nil
}

func makeBlockDay(b *types.Block) uint32 {
	startTm := time.Date(1900, time.Month(1), 1, 0, 0, 0, 0, time.UTC)
	tm := time.Unix(0, int64(b.Header.Timestamp))
	du := tm.Sub(startTm)
	dayDiff := du / time.Hour / 24
	days := uint32(dayDiff)
	return days
}

func (t *TxSearch) saveRewardEvent(en *ctypes.Event, batch *leveldb.Batch, days uint32) {
	mp := map[common.Address][]byte{}
	if err := types.UnmarshalAddressBytesMap(en.Result, mp); err != nil {
		return
	}
	for cont, bs := range mp {
		ma := map[common.Address]*amount.Amount{}
		if err := types.UnmarshalAddressAmountMap(bs, ma); err != nil {
			return
		}
		for rewarder, am := range ma {
			t.saveReward(batch, cont, rewarder, am)
			t.saveDailyReward(batch, cont, rewarder, am, days)
		}
	}
}

func (t *TxSearch) saveReward(batch *leveldb.Batch, cont, rewarder common.Address, arg *amount.Amount) {
	key := make([]byte, 41)
	key[0] = tagEventReward
	copy(key[1:], cont[:])
	copy(key[21:], rewarder[:])
	bs, err := t.db.Get(key, nil)
	if err != nil && err != leveldb.ErrNotFound {
		return
	}
	am := amount.NewAmountFromBytes(bs)
	am.Int.Add(am.Int, arg.Int)
	batch.Put(key, am.Bytes())
}

func (t *TxSearch) saveDailyReward(batch *leveldb.Batch, cont, rewarder common.Address, arg *amount.Amount, days uint32) {
	key := make([]byte, 45)
	key[0] = tagDailyReward
	copy(key[1:], cont[:])
	copy(key[21:], rewarder[:])

	bs := make([]byte, 4)
	binary.BigEndian.PutUint32(bs, days)
	copy(key[41:], bs[:])

	bs, err := t.db.Get(key, nil)
	if err != nil && err != leveldb.ErrNotFound {
		return
	}
	am := amount.NewAmountFromBytes(bs)
	am.Int.Add(am.Int, arg.Int)
	batch.Put(key, am.Bytes())
}

func (t *TxSearch) saveTx(indexMap map[addrIndexKey]uint64, index41Map map[addr41IndexKey]uint64, batch *leveldb.Batch, tx *types.Transaction, height uint32, TXID []byte) {
	empty := common.Address{}

	t.Push(indexMap, batch, addrKey(tagID, empty[:]), TXID, tx.From, tx.To, tx.Method)
	t.Push(indexMap, batch, addrKey(tagDefault, tx.From[:]), tx.Method, TXID, tx.To)
	t.Push(indexMap, batch, addrKey(tagDefault, tx.To[:]), tx.Method, TXID, tx.From)
	t._saveTx(indexMap, index41Map, batch, tx, height, TXID)
}

func (t *TxSearch) _saveTx(indexMap map[addrIndexKey]uint64, index41Map map[addr41IndexKey]uint64, batch *leveldb.Batch, tx *types.Transaction, height uint32, TXID []byte) {
	TxTo, method, arg, err := types.TxArg(t.cn.NewContext(), tx)
	if err != nil {
		return
	}

	hbs := make([]byte, 4)
	binary.BigEndian.PutUint32(hbs, height)

	switch strings.ToLower(method) {
	case "transfer":
		to := toAddress(arg, 0)
		am := toAmount(arg, 1)
		t.Push(indexMap, batch, addrKey(tagAddress, tx.From[:]), tx.Method, TXID, TxTo, uint8(0), to, am)
		t.Push(indexMap, batch, addrKey(tagAddress, to[:]), tx.Method, TXID, TxTo, uint8(1), tx.From, am)

		t.Push41(index41Map, batch, addr41Key(tagTransfer, TxTo, tx.From), tx.Method, TXID, uint8(0), to, am)
		t.Push41(index41Map, batch, addr41Key(tagTransfer, TxTo, to), tx.Method, TXID, uint8(1), tx.From, am)
	case "transferfrom":
		From := toAddress(arg, 0)
		To := toAddress(arg, 1)
		am := toAmount(arg, 2)
		t.Push(indexMap, batch, addrKey(tagAddress, From[:]), tx.Method, TXID, TxTo, uint8(0), To, am)
		t.Push(indexMap, batch, addrKey(tagAddress, To[:]), tx.Method, TXID, TxTo, uint8(1), From, am)

		t.Push41(index41Map, batch, addr41Key(tagTransfer, TxTo, From), tx.Method, TXID, uint8(0), To, am)
		t.Push41(index41Map, batch, addr41Key(tagTransfer, TxTo, To), tx.Method, TXID, uint8(1), From, am)
	case "approve":
		t.Push(indexMap, batch, addrKey(tagAddress, tx.From[:]), tx.Method, TXID, TxTo)
	case "createalpha", "createsigma", "createomega":
		t.Push(indexMap, batch, addrKey(tagAddress, tx.From[:]), tx.Method, TXID)
	case "revoke":
		t.Push(indexMap, batch, addrKey(tagAddress, tx.From[:]), tx.Method, TXID)
	case "stake":
		HyperAddress := toAddress(arg, 0)
		t.Push(indexMap, batch, addrKey(tagAddress, tx.From[:]), tx.Method, TXID, HyperAddress)
	case "unstake":
		HyperAddress := toAddress(arg, 0)
		t.Push(indexMap, batch, addrKey(tagAddress, tx.From[:]), tx.Method, TXID, HyperAddress)
	case "tokenin", "tokenindexin":
		to := toAddress(arg, 2)
		am := toAmount(arg, 3)

		t.Push(indexMap, batch, addrKey(tagAddress, tx.From[:]), tx.Method, TXID, TxTo, uint8(0), to, am)
		t.Push(indexMap, batch, addrKey(tagAddress, to[:]), tx.Method, TXID, TxTo, uint8(1), tx.From, am)

		t.Push41(index41Map, batch, addr41Key(tagTransfer, *t.st.MainToken(), tx.From), tx.Method, TXID, uint8(0), to, am)
		t.Push41(index41Map, batch, addr41Key(tagTransfer, *t.st.MainToken(), to), tx.Method, TXID, uint8(1), tx.From, am)
	case "tokenleave":
		CoinTXID := toString(arg, 0)
		ERC20TXID := toString(arg, 1)
		Platform := toString(arg, 2)
		t.Push(indexMap, batch, addrKey(tagTokenLeave, hbs), TXID, CoinTXID, ERC20TXID, Platform)
	case "tokenout":
		Platform := toString(arg, 0)
		withdrawAddress := toAddress(arg, 1)
		am := toAmount(arg, 2)

		t.Push(indexMap, batch, addrKey(tagTokenOut, hbs), TXID, tx.From, Platform, withdrawAddress, am.Int)

		t.Push(indexMap, batch, addrKey(tagAddress, tx.From[:]), tx.Method, TXID)
		t.Push41(index41Map, batch, addr41Key(tagTransfer, *t.st.MainToken(), tx.From), tx.Method, TXID, uint8(0), withdrawAddress, am)
	case "sendtogateway":
		t.Push(indexMap, batch, addrKey(tagBridge, TxTo[:], hbs), TXID)
		token := toAddress(arg, 0)
		am := toAmount(arg, 1)
		t.Push41(index41Map, batch, addr41Key(tagTransfer, token, tx.From), tx.Method, TXID, uint8(0), tx.To, am)
	case "sendfromgateway":
		t.Push(indexMap, batch, addrKey(tagBridge, TxTo[:], hbs), TXID)
		token := toAddress(arg, 0)
		to := toAddress(arg, 1)
		am := toAmount(arg, 2)
		t.Push41(index41Map, batch, addr41Key(tagTransfer, token, to), tx.Method, TXID, uint8(1), tx.From, am)
	case "sendtorouterfromgateway":
		t.Push(indexMap, batch, addrKey(tagBridge, TxTo[:], hbs), TXID)
	default:
		t.Push(indexMap, batch, addrKey(tagAddress, tx.From[:]), tx.Method, TXID)
	}
}

func toString(is []interface{}, index int) (res string) {
	if len(is) <= index {
		return
	}

	str := is[index]
	switch s := str.(type) {
	case string:
		res = s
	default:
		res = fmt.Sprintf("%v", s)
	}
	return
}

func toAddress(is []interface{}, index int) (res common.Address) {
	if len(is) <= index {
		return
	}

	iaddr := is[index]
	switch addr := iaddr.(type) {
	case common.Address:
		res = addr
	case string:
		res = common.HexToAddress(addr)
	case *big.Int:
		res = common.BytesToAddress(addr.Bytes())
	}
	return
}

func toAmount(is []interface{}, index int) (am *amount.Amount) {
	if len(is) <= index {
		return
	}

	i := is[index]
	switch amt := i.(type) {
	case *amount.Amount:
		am = amt
	case *big.Int:
		am = &amount.Amount{Int: amt}
	}
	return am
}

func addrKey(tag byte, addrs ...[]byte) []byte {
	bs := []byte{}
	bs = append(bs, tag)
	for _, v := range addrs {
		bs = append(bs, v...)
	}
	return bs
}

func addr41Key(tag byte, addr, addr2 common.Address) []byte {
	bs := make([]byte, (common.AddressLength*2)+1)
	bs[0] = tag
	copy(bs[1:], addr[:])
	copy(bs[21:], addr2[:])
	return bs
}

func (t *TxSearch) Push(indexMap map[addrIndexKey]uint64, batch *leveldb.Batch, tag []byte, args ...interface{}) {
	value := bin.TypeWriteAll(args...)
	var aik addrIndexKey
	copy(aik[:], tag)
	var index uint64
	var ok bool
	if index, ok = indexMap[aik]; !ok {
		bs, _ := t.db.Get(aik[:], nil)
		if len(bs) != 8 {
			bs = make([]byte, 8)
		}
		index = bin.Uint64(bs)
	}
	index++
	indexMap[aik] = index
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, index)
	batch.Put(append(tag, bs...), value)
}

func (t *TxSearch) Push41(indexMap map[addr41IndexKey]uint64, batch *leveldb.Batch, tag []byte, args ...interface{}) {
	value := bin.TypeWriteAll(args...)
	var aik addr41IndexKey
	copy(aik[:], tag)
	var index uint64
	var ok bool
	if index, ok = indexMap[aik]; !ok {
		bs, _ := t.db.Get(aik[:], nil)
		if len(bs) != 8 {
			bs = make([]byte, 8)
		}
		index = bin.Uint64(bs)
	}
	index++
	indexMap[aik] = index
	bs := make([]byte, 8)
	binary.BigEndian.PutUint64(bs, index)
	batch.Put(append(tag, bs...), value)
}
