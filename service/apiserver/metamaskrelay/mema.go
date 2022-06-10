package metamaskrelay

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ecommon "github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/extern/txparser"
	"github.com/meverselabs/meverse/service/apiserver"
	"github.com/meverselabs/meverse/service/apiserver/viewchain"
	"github.com/meverselabs/meverse/service/txsearch"
)

type INode interface {
	AddTx(tx *types.Transaction, sig common.Signature) error
}

const (
	logsBloom = "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
)

type metamaskRelay struct {
	api     *apiserver.APIServer
	chainID *big.Int
	ts      txsearch.ITxSearch
	cn      *chain.Chain
	nd      INode
}

func NewMetamaskRelay(api *apiserver.APIServer, ts txsearch.ITxSearch, cn *chain.Chain, nd INode) {
	m := &metamaskRelay{
		api:     api,
		chainID: cn.Provider().ChainID(),
		ts:      ts,
		cn:      cn,
		nd:      nd,
	}

	s, err := m.api.JRPC("eth")
	if err != nil {
		panic(err)
	}

	// chainID := "0xffff"
	s.Set("net_version", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return fmt.Sprintf("0x%x", m.chainID.Uint64()), nil
	})
	s.Set("eth_chainId", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return fmt.Sprintf("0x%x", m.chainID.Uint64()), nil
	})
	s.Set("eth_blockNumber", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		height := cn.Provider().Height()
		log.Println("eth_blockNumber", height, fmt.Sprintf("0x%x", height))
		return fmt.Sprintf("0x%x", height), nil
	})
	s.Set("eth_getBlockByNumber", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		height, _ := arg.String(0)
		txFull, _ := arg.String(1)
		log.Println("eth_getBlockByNumber", height, txFull)
		cheight := cn.Provider().Height()
		var hei uint64
		if height == "latest" {
			hei = uint64(cheight)
		} else {
			cleaned := strings.Replace(height, "0x", "", -1)
			hei, _ = strconv.ParseUint(cleaned, 16, 64)
		}
		if uint32(hei) > cheight {
			hei = 1
		}
		return m.returnMemaBlock(hei, txFull == "true")
	})
	s.Set("eth_getBlockByHash", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		bhash, _ := arg.String(0)
		txFull, _ := arg.String(1)
		log.Println("eth_getBlockByHash", bhash)

		cleaned := strings.Replace(bhash, "0x", "", -1)

		bs, err := hex.DecodeString(cleaned)
		if err != nil {
			return nil, err
		}
		var hs hash.Hash256
		if len(hs) != len(bs) {
			return nil, errors.New("invalid hash length")
		}
		copy(hs[:], bs[:])

		hei, err := ts.BlockHeight(hs)
		if err != nil {
			return nil, err
		}
		if uint32(hei) > cn.Provider().Height() {
			hei = 1
		}
		return m.returnMemaBlock(uint64(hei), txFull == "true")
	})
	s.Set("eth_getBalance", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		addrStr, _ := arg.String(0)
		mainaddr := cn.NewContext().MainToken()
		if mainaddr == nil {
			return "0x0", nil
		}

		addr, err := common.ParseAddress(addrStr)
		if err != nil {
			return nil, err
		}

		rv, err := m._getTokenBalanceOf(*mainaddr, addr)
		if err != nil {
			return nil, err
		}
		return rv, nil
	})
	s.Set("eth_gasPrice", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		// return "0x3B9ACA00", nil
		return "0xE8D4A51000", nil
	})

	s.Set("eth_estimateGas", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		// return "0xcf08", nil 79500
		// return "0x1DCD6500", nil
		return "0x0186A0", nil
	})

	s.Set("eth_getCode", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		addr, err := arg.String(0)
		if err != nil {
			return "0x", fmt.Errorf("invalid params %v", addr)
		}
		contAddr := common.HexToAddress(addr)
		if cn.NewContext().IsContract(contAddr) {
			head := "6080604052348015600f57600080fd5b50603580601d6000396000f3fe6080604052600080fdfea165627a7a72305820"
			tail := "6080604052348015600f57600080fd5b50603580601d6000396000f3fe6080604052600080fdfea165627a7a72305820"
			hexString := "0x" + head + addr + tail
			log.Println(hexString)
			return hexString, nil
		}
		return "0x", nil
	})

	s.Set("eth_sendRawTransaction", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		rlp, _ := arg.String(0)
		rlp1, _ := arg.String(1)
		rlp2, _ := arg.String(2)
		log.Println("eth_sendRawTransaction", rlp, rlp1, rlp2)

		tx, sig, err := m.TransmuteTx(rlp)
		if err != nil {
			return nil, err
		}

		TxHash := tx.Hash()
		pubkey, err := common.RecoverPubkey(tx.ChainID, TxHash, sig)
		if err != nil {
			return nil, err
		}
		From := pubkey.Address()

		ctx := m.cn.NewContext()
		n := ctx.Snapshot()
		txid := types.TransactionID(ctx.TargetHeight(), 0)
		if tx.To == common.ZeroAddr {
			_, err = m.cn.ExecuteTransaction(ctx, tx, txid)
		} else {
			err = chain.ExecuteContractTx(ctx, tx, From, "000000000000")
		}
		// if err != nil && !strings.Contains(err.Error(), "invalid signer sequence siger seq") {
		if err != nil {
			log.Printf("%+v\n", err)
			return nil, err
		}
		ctx.Revert(n)

		err = m.nd.AddTx(tx, sig)
		if err != nil {
			return nil, err
		}

		// m.ts.AddTXPool(hash) TODO
		log.Println("eth_sendRawTransaction done", tx.Hash().String(), rlp)
		return tx.Hash().String(), nil
	})

	s.Set("eth_getTransactionCount", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		addrStr, _ := arg.String(0)
		addr, err := common.ParseAddress(addrStr)
		if err != nil {
			return nil, err
		}
		seq := m.cn.NewContext().AddrSeq(addr)
		return "0x" + strconv.FormatUint(seq, 16), nil
	})

	s.Set("eth_getTransactionReceipt", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		txhash, err := arg.String(0)
		if err != nil {
			return nil, err
		}
		log.Println("eth_getTransactionReceipt", txhash)
		hs := hash.HexToHash(txhash)

		TxID, err := ts.TxIndex(hs)
		if err != nil {
			return nil, err
		}

		if TxID.Err != nil {
			b, err := m.cn.Provider().Block(TxID.Height)
			if err != nil {
				return nil, err
			}

			bHash := bin.MustWriterToHash(&b.Header)

			return map[string]interface{}{
				"transactionHash":   hs,
				"transactionIndex":  fmt.Sprintf("0x%x", TxID.Index+1),
				"blockNumber":       fmt.Sprintf("0x%x", TxID.Height),
				"blockHash":         bHash.String(),
				"cumulativeGasUsed": "0x0",
				"gasUsed":           "0x0",
				"contractAddress":   nil, // or null, if none was created
				"logs":              []string{},
				"logsBloom":         logsBloom,
				"status":            "0x0", //TODO 성공 1 실패 0
			}, nil
		}

		b, err := m.cn.Provider().Block(TxID.Height)
		if err != nil {
			return nil, err
		}
		if int(TxID.Index) >= len(b.Body.Transactions) {
			return nil, errors.New("invalid txhash")
		}
		tx := b.Body.Transactions[TxID.Index]

		bHash := bin.MustWriterToHash(&b.Header)
		m := map[string]interface{}{
			// "cumulativeGasUsed": "0x1f4b698",
			// "gasUsed":           "0x6a20",
			"blockHash":         bHash.String(),
			"blockNumber":       fmt.Sprintf("0x%x", TxID.Height),
			"transactionHash":   tx.Hash(),
			"transactionIndex":  fmt.Sprintf("0x%x", TxID.Index),
			"from":              tx.From.String(),
			"to":                tx.To.String(),
			"cumulativeGasUsed": "0x1f4b698",
			"gasUsed":           "0x1f4b698",
			"effectiveGasPrice": "0x71e0e496c",
			"contractAddress":   nil, // or null, if none was created
			"logs":              []string{},
			"logsBloom":         logsBloom,
			"type":              "0x2",
			"status":            "0x1", //TODO 성공 1 실패 0
			"data":              map[string]string{"test": "success"},
		}
		return m, nil
	})

	s.Set("eth_getTransactionByHash", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		txhash, err := arg.String(0)
		if err != nil {
			return nil, err
		}
		hash.HexToHash(txhash)
		hs := hash.HexToHash(txhash)
		TxID, err := ts.TxIndex(hs)
		if err != nil {
			return nil, err
		}
		if TxID.Err != nil {
			return nil, TxID.Err
		}

		return getTransaction(m, TxID.Height, TxID.Index)
	})
	s.Set("eth_getBlockTransactionCountByNumber", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		heightHex, err := arg.String(0)
		if err != nil {
			return nil, err
		}
		heightHex = strings.ReplaceAll(heightHex, "0x", "")
		height, err := strconv.ParseUint(heightHex, 16, 64)
		if err != nil {
			return nil, err
		}

		b, err := m.cn.Provider().Block(uint32(height))
		if err != nil {
			return nil, err
		}
		return fmt.Sprintf("0x%x", len(b.Body.Transactions)), nil
	})
	s.Set("eth_getTransactionByBlockNumberAndIndex", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		heightHex, err := arg.String(0)
		if err != nil {
			return nil, err
		}
		heightHex = strings.ReplaceAll(heightHex, "0x", "")
		height64, err := strconv.ParseUint(heightHex, 16, 64)
		if err != nil {
			return nil, err
		}
		height := uint32(height64)
		indexHex, err := arg.String(1)
		if err != nil {
			return nil, err
		}
		indexHex = strings.ReplaceAll(indexHex, "0x", "")
		index64, err := strconv.ParseUint(indexHex, 16, 16)
		if err != nil {
			return nil, err
		}
		index := uint16(index64)

		return getTransaction(m, height, index)
	})

	s.Set("eth_call", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		param, _ := arg.Map(0)
		h, _ := arg.String(1)
		var to string
		var data string
		var from string
		if tol, ok := param["to"].(string); ok {
			to = tol
		}
		if datal, ok := param["data"].(string); ok {
			data = datal
		}
		if froml, ok := param["from"].(string); ok {
			from = froml
		}
		return m.ethCall(h, to, data, from)
	})
}

func getTransaction(m *metamaskRelay, height uint32, index uint16) (interface{}, error) {
	b, err := m.cn.Provider().Block(height)
	if err != nil {
		return nil, err
	}
	bHash := bin.MustWriterToHash(&b.Header)

	if int(index) >= len(b.Body.Transactions) {
		return nil, errors.New("invalid txhash")
	}
	tx := b.Body.Transactions[index]
	sig := b.Body.TransactionSignatures[index]

	to := tx.To.String()
	var (
		value string = "0x0"
		nonce string
		input string
	)

	mainToken := m.cn.NewContext().MainToken()
	mto, amt, err := getMainTokenSend(*mainToken, tx)
	if err == nil {
		to = mto.String()
		value = fmt.Sprintf("0x%v", hex.EncodeToString(amt.Bytes()))
	}
	if tx.IsEtherType {
		etx, _, err := txparser.EthTxFromRLP(tx.Args)
		if err != nil {
			return nil, err
		}
		nonce = fmt.Sprintf("0x%x", etx.Nonce())
		input = fmt.Sprintf("0x%v", hex.EncodeToString(etx.Data()))
	} else {
		nonce = fmt.Sprintf("0x%x", tx.Seq)
		input = fmt.Sprintf("0x%v%v", hex.EncodeToString([]byte(tx.Method)), hex.EncodeToString(tx.Args))
	}

	return map[string]interface{}{
		"blockHash":        bHash.String(),
		"blockNumber":      fmt.Sprintf("0x%x", height),
		"from":             tx.From.String(),
		"gas":              "0x1f4b698",
		"gasPrice":         "0x71e0e496c",
		"hash":             tx.Hash().String(),
		"input":            input,
		"nonce":            nonce,
		"to":               to,
		"transactionIndex": fmt.Sprintf("0x%x", index),
		"value":            value,
		"v":                "0x" + hex.EncodeToString(sig[64:]),
		"r":                "0x" + hex.EncodeToString(sig[:32]),
		"s":                "0x" + hex.EncodeToString(sig[32:64]),
	}, nil
}

func getMainTokenSend(MainToken common.Address, tx *types.Transaction) (to common.Address, amt *big.Int, err error) {
	if tx.IsEtherType {
		var etx *etypes.Transaction
		etx, _, err = txparser.EthTxFromRLP(tx.Args)
		if err != nil {
			return
		}
		eData := etx.Data()
		if etx.Value().Cmp(amount.ZeroCoin.Int) > 0 && tx.To != MainToken && len(eData) == 0 {
			to = tx.To
			amt = etx.Value()
			return
		} else if len(eData) > 0 && strings.EqualFold(etx.To().String(), MainToken.String()) {
			var m abi.Method
			for _, m = range txparser.FuncSigs[hex.EncodeToString(etx.Data()[:4])] {
				break
			}
			if strings.ToLower(m.Name) == "transfer" {
				var data []interface{}
				data, err = txparser.Inputs(eData)
				to, amt, err = getTransferParam(data, err)
				return
			}
		}
	} else if strings.EqualFold(tx.To.String(), MainToken.String()) {
		data, _err := bin.TypeReadAll(tx.Args, -1)
		to, amt, err = getTransferParam(data, _err)
		return
	}
	err = errors.New("is not maintoken transfer")
	return
}

func getTransferParam(data []interface{}, _err error) (to common.Address, amt *big.Int, err error) {
	if err != nil {
		return common.Address{}, nil, err
	}
	if len(data) != 2 {
		return common.Address{}, nil, errors.New("invalid param")
	}
	var ok bool
	if to, ok = data[0].(common.Address); !ok {
		return common.Address{}, nil, errors.New("invalid address param")
	}
	if am, ok := data[1].(*amount.Amount); !ok {
		return common.Address{}, nil, errors.New("invalid amount param")
	} else {
		amt = am.Int
	}
	return
}

func (m *metamaskRelay) returnMemaBlock(hei uint64, fullTx bool) (interface{}, error) {
	b, err := m.cn.Provider().Block(uint32(hei))
	if err != nil {
		log.Printf("eth_getBlockByNumber err %v %+v\n", uint32(hei), err)
		return nil, err
	}

	bHash := bin.MustWriterToHash(&b.Header)

	txs := []string{}
	for _, tx := range b.Body.Transactions {
		txs = append(txs, tx.Hash().String())
	}

	return map[string]interface{}{
		"gasUsed":          "0x1c9c380",
		"gasLimit":         "0x1c9c380",
		"hash":             bHash.String(),
		"logsBloom":        logsBloom,
		"number":           fmt.Sprintf("0x%x", b.Header.Height),
		"parentHash":       b.Header.PrevHash.String(),
		"size":             fmt.Sprintf("0x%x", len(b.Body.Transactions)),
		"timestamp":        fmt.Sprintf("0x%x", b.Header.Timestamp/1000),
		"transactions":     txs,
		"transactionsRoot": b.Header.LevelRootHash.String(),
	}, nil
}

func (m *metamaskRelay) TransmuteTx(rlp string) (*types.Transaction, common.Signature, error) {
	rlp = strings.Replace(rlp, "0x", "", -1)
	rlpBytes, err := hex.DecodeString(rlp)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	etx, sig, err := txparser.EthTxFromRLP(rlpBytes)
	if err != nil {
		return nil, nil, err
	}

	isInvokeable := false
	{
		ctx := m.cn.NewContext()
		var cont interface{}
		cont, err = ctx.Contract(*etx.To())
		if err == nil {
			if _, ok := cont.(types.InvokeableContract); ok {
				isInvokeable = true
			}
		}
	}

	var method string
	if len(etx.Data()) >= 4 {
		var m abi.Method
		for _, m = range txparser.FuncSigs[hex.EncodeToString(etx.Data()[:4])] {
			break
		}
		if m.Name == "" {
			return nil, nil, errors.New("not exist abi")
		}
		if isInvokeable {
			method = m.Name
		} else {
			method = strings.ToUpper(string(m.Name[0])) + m.Name[1:]
		}
	}
	if method == "" {
		var m abi.Method
		for _, m = range txparser.FuncSigs["transfer(address,uint256)"] {
			break
		}
		name := m.Name
		if isInvokeable {
			method = name
		} else {
			method = strings.ToUpper(string(name[0])) + name[1:]
		}
	}

	gp := etx.GasPrice()
	if gp == nil || len(gp.Bytes()) == 0 {
		gp = etx.GasFeeCap()
	}

	tx := &types.Transaction{
		ChainID:     m.cn.Provider().ChainID(),
		Timestamp:   uint64(time.Now().UnixNano()),
		To:          *etx.To(),
		Method:      method,
		Args:        rlpBytes,
		Seq:         etx.Nonce(),
		UseSeq:      true,
		IsEtherType: true,
		GasPrice:    gp,
	}
	return tx, sig, nil
}

func (m *metamaskRelay) getTokenContract(conAddr common.Address) (*token.TokenContract, types.ContractLoader, error) {
	ctx := m.cn.NewContext()
	v, err := ctx.Contract(conAddr)
	if err != nil {
		return nil, nil, err
	}
	if cont, ok := v.(*token.TokenContract); ok {
		cc := ctx.ContractLoader(cont.Address())
		return cont, cc, nil
	}
	return nil, nil, errors.New("not match contract")
}

func (m *metamaskRelay) _getTokenBalanceOf(to common.Address, addr common.Address) (string, error) {
	cont, cc, err := m.getTokenContract(to)
	if err != nil {
		return "0x0", err
	}

	am := cont.BalanceOf(cc, addr)

	rv := "0x" + hex.EncodeToString(am.Bytes())
	if rv == "0x" {
		rv = "0x0"
	}
	return rv, nil
}

func (m *metamaskRelay) ethCall(height, to, data, from string) (interface{}, error) {
	if len(data) < 10 {
		log.Println("ErrInvalidData len:", len(data))
		return nil, errors.WithStack(ErrInvalidData)
	}
	toAddr, err := common.ParseAddress(to)
	if err != nil {
		return nil, err
	}
	if strings.Index(data, "0x") == 0 {
		data = data[2:]
	}
	abiMs := txparser.FuncSigs[data[:8]]
	var _err error
	for _, abiM := range abiMs {
		if abiM.StateMutability == "view" {
			bs, err := hex.DecodeString(data)
			if err != nil {
				_err = err
				continue
			}
			obj, err := txparser.Inputs(bs)
			if err != nil {
				_err = err
				continue
			}
			caller := viewchain.NewViewCaller(m.cn)
			output, err := caller.Execute(toAddr, from, abiM.Name, obj)
			if err != nil {
				_err = fmt.Errorf("%v call %v method %v", err.Error(), toAddr, abiM.Name)
				continue
			}

			bs, err = txparser.Outputs(abiM.Sig, output)
			if err != nil {
				_err = err
				continue
			}
			return "0x" + hex.EncodeToString(bs), nil
		}
	}
	if _err != nil {
		return nil, _err
	}
	return nil, errors.New("func not found")
}

func makeStringResponse(str string) string {
	var t []byte
	{
		buf := new(bytes.Buffer)
		var num uint8 = 32
		err := binary.Write(buf, binary.LittleEndian, num)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
		}
		t = ecommon.LeftPadBytes(buf.Bytes(), 32)
	}
	var slen []byte
	{
		buf := new(bytes.Buffer)
		var strlen uint8 = uint8(len(str))
		err := binary.Write(buf, binary.LittleEndian, strlen)
		if err != nil {
			fmt.Println("binary.Write failed:", err)
		}
		slen = ecommon.LeftPadBytes(buf.Bytes(), 32)
	}

	s := ecommon.RightPadBytes([]byte(str), 32)

	var data []byte
	data = append(data, t...)
	data = append(data, slen...)
	data = append(data, s...)

	return fmt.Sprintf("0x%x", data)
}
