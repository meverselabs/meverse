package metamaskrelay

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ecommon "github.com/ethereum/go-ethereum/common"
	ecore "github.com/ethereum/go-ethereum/core"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/contract/token"
	"github.com/meverselabs/meverse/core/chain"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/ethereum/core"
	mcore "github.com/meverselabs/meverse/ethereum/core"
	"github.com/meverselabs/meverse/ethereum/core/defaultevm"
	mtypes "github.com/meverselabs/meverse/ethereum/core/types"
	"github.com/meverselabs/meverse/extern/txparser"
	"github.com/meverselabs/meverse/service/apiserver"
	"github.com/meverselabs/meverse/service/apiserver/viewchain"
	"github.com/meverselabs/meverse/service/bloomservice"
	"github.com/meverselabs/meverse/service/txsearch/itxsearch"
)

type INode interface {
	AddTx(tx *types.Transaction, sig common.Signature) error
}

var (
	logsBloom = "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
)

type metamaskRelay struct {
	api     *apiserver.APIServer
	chainID *big.Int
	ts      itxsearch.ITxSearch
	bs      *bloomservice.BloomBitService
	cn      *chain.Chain
	nd      INode
}

func NewMetamaskRelay(api *apiserver.APIServer, ts itxsearch.ITxSearch, bs *bloomservice.BloomBitService, cn *chain.Chain, nd INode) {
	m := &metamaskRelay{
		api:     api,
		chainID: cn.Provider().ChainID(),
		ts:      ts,
		bs:      bs,
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
	s.Set("web3_clientVersion", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		//return fmt.Sprintf("Meverse/v2.0"), nil
		return fmt.Sprintf("Geth/v1.10.17-stable/linux-amd64/go1.18.3"), nil
	})
	s.Set("eth_feeHistory", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		oldestBlock, _ := arg.String(1)
		if oldestBlock == "latest" || oldestBlock == "pending" {
			cheight := cn.Provider().Height()
			oldestBlock = strconv.FormatUint(uint64(cheight), 16)
		}
		return map[string]interface{}{
			"oldestBlock":   oldestBlock,
			"baseFeePerGas": []string{m.getBaseFeePerGas(), m.getBaseFeePerGas()},
			"gasUsedRatio":  []float32{0.5},
			"reward":        [][]string{{"0x0"}},
		}, nil
	})

	s.Set("eth_chainId", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return fmt.Sprintf("0x%x", m.chainID.Uint64()), nil
	})
	s.Set("eth_blockNumber", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		height := cn.Provider().Height()
		return fmt.Sprintf("0x%x", height), nil
	})
	s.Set("eth_getBlockByNumber", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		height, _ := arg.String(0)
		txFull, _ := arg.String(1)
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
		switch m.Version() {
		case 0, 1:
			return "0xE8D4A51000", nil
		default:
			bs := m.cn.NewContext().BasicFee().Bytes()
			return fmt.Sprintf("0x%s", hex.EncodeToString(bs)), nil
		}
	})

	s.Set("eth_estimateGas", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		switch m.Version() {
		case 0, 1:
			//return "0x0186A0", nil
			return "0x1DCD6500", nil
		default:
			//log.Println("arg in eth_estimateGas", arg)
			argMap, err := arg.Map(0)
			if err != nil {
				return "0x1DCD6500", nil // 50000000
			}

			var to, from, data string
			if tol, ok := argMap["to"].(string); ok {
				to = tol
			}

			if datal, ok := argMap["data"].(string); ok {
				data = datal
			}
			if froml, ok := argMap["from"].(string); ok {
				from = froml
			}

			if data == "" {
				if value, ok := argMap["value"].(string); ok {
					data = fmt.Sprintf("0xa9059cbb%064v%064v", strings.Replace(to, "0x", "", -1), strings.Replace(value, "0x", "", -1))
					to = m.cn.NewContext().MainToken().String()
				} else {
					return 0, errors.New("invalid value")
				}
			}
			//_, gas, err := m.ethCall(from, to, data)

			//return gas, err

			gasDefault := "0x1DCD6500"
			_, _, err = m.ethCall(from, to, data)
			return gasDefault, err
		}

		// return "0xcf08", nil 79500
		//return "0x1d023", nil // 0x1d023 = decimal 118819
	})

	s.Set("eth_getCode", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		addr, err := arg.String(0)
		if err != nil {
			return "0x", fmt.Errorf("invalid params %v", addr)
		}
		contAddr := common.HexToAddress(addr)
		ctx := cn.NewContext()

		if ctx.IsContract(contAddr) {
			head := "6080604052348015600f57600080fd5b50603580601d6000396000f3fe6080604052600080fdfea165627a7a72305820"
			tail := "6080604052348015600f57600080fd5b50603580601d6000396000f3fe6080604052600080fdfea165627a7a72305820"
			hexString := "0x" + head + addr + tail
			return hexString, nil
		} else {
			code := types.NewStateDB(ctx).GetCode(contAddr)
			if len(code) == 0 {
				return "0x", nil
			} else {
				return hex.EncodeToString(code), nil
				// return code, nil
			}
		}
	})

	s.Set("eth_sendRawTransaction", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {

		rlp, _ := arg.String(0)
		// rlp1, _ := arg.String(1)
		// rlp2, _ := arg.String(2)
		//log.Println("eth_sendRawTransaction", "rlp", rlp, "rlp1", rlp1, "rlp2", rlp2)

		return m.transactionHash(rlp)
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
		//log.Println("eth_getTransactionReceipt", txhash)
		hs := hash.HexToHash(txhash)

		TxID, err := ts.TxIndex(hs)
		if err != nil {
			if err.Error() == "leveldb: not found" {
				return nil, nil
			}
			return nil, err
		}
		b, err := m.cn.Provider().Block(TxID.Height)
		if err != nil {
			return nil, err
		}
		bHash := bin.MustWriterToHash(&b.Header)

		if TxID.Err != nil {
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

		if int(TxID.Index) >= len(b.Body.Transactions) {
			return nil, errors.New("invalid txhash")
		}

		tx := b.Body.Transactions[TxID.Index]
		return getReceipt(tx, b, TxID, m, bHash)

	})

	s.Set("eth_getTransactionByHash", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		txhash, err := arg.String(0)
		if err != nil {
			return nil, err
		}
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
		result, _, err := m.ethCall(from, to, data)

		return result, err
	})

	s.Set("eth_getLogs", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {

		filterMap, err := arg.Map(0)
		if err != nil {
			return nil, err
		}
		return bloomservice.FilterLogs(m.cn, m.ts, m.bs, bloomservice.ToFilter(filterMap))

	})

	s.Set("web3_clientVersion", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return viewchain.GetVersion(), nil
	})
	s.Set("eth_clientVersion", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return viewchain.GetVersion(), nil
	})
}

func getReceipt(tx *types.Transaction, b *types.Block, TxID itxsearch.TxID, m *metamaskRelay, bHash ecommon.Hash) (map[string]interface{}, error) {
	result := map[string]interface{}{
		// "cumulativeGasUsed": "0x1f4b698",
		// "gasUsed":           "0x6a20",
		"blockHash":         bHash.String(),
		"blockNumber":       fmt.Sprintf("0x%x", TxID.Height),
		"transactionHash":   tx.HashSig(),
		"transactionIndex":  fmt.Sprintf("0x%x", TxID.Index),
		"from":              tx.From.String(),
		"to":                tx.To.String(),
		"cumulativeGasUsed": "0x1f4b698",
		"gasUsed":           "0x1f4b698",
		"effectiveGasPrice": "0x71e0e496c",
		"contractAddress":   nil, // or null, if none was created
		"type":              "0x2",
		"status":            "0x1", //TODO 성공 1 실패 0
		"data":              map[string]string{},
	}

	if tx.VmType != types.Evm {
		evs, err := bloomservice.FindTransactionsEvents(b.Body.Transactions, b.Body.Events, int(TxID.Index))
		if err != nil {
			return nil, err
		}
		blm, err := bloomservice.CreateEventBloom(m.cn.NewContext(), evs)
		if err != nil {
			return nil, err
		}

		logs, err := bloomservice.EventsToLogs(m.cn, &b.Header, tx, evs, int(TxID.Index))
		if err != nil {
			return nil, err
		}
		result["logs"] = logs
		result["logsBloom"] = hex.EncodeToString(blm[:])

	} else {
		etx := new(etypes.Transaction)
		if err := etx.UnmarshalBinary(tx.Args); err != nil {
			return nil, err
		}
		eHash := etx.Hash()
		receipts, err := m.cn.Provider().Receipts(TxID.Height)
		if err != nil {
			return nil, err
		}
		if len(receipts) <= int(TxID.Index) {
			return nil, nil
		}
		receipt := receipts[TxID.Index]
		signer := mtypes.MakeSigner(m.cn.Provider().ChainID(), TxID.Height)
		if err := receipts.DeriveReceiptFields(bHash, uint64(TxID.Height), TxID.Index, etx, signer); err != nil {
			return nil, err
		}

		logs := []*etypes.Log{}
		if len(receipt.Logs) >= 0 {
			logs = append(logs, receipt.Logs...)
		}
		bloom := etypes.CreateBloom(etypes.Receipts{receipt})
		//ContractAddresss : 하단 참조

		var to interface{}
		if etx.To() != nil {
			to = etx.To().String()
		} else {
			to = nil
		}

		result["transactionHash"] = eHash.String()
		result["to"] = to
		result["cumulativeGasUsed"] = fmt.Sprintf("0x%x", receipt.CumulativeGasUsed)
		result["gasUsed"] = fmt.Sprintf("0x%x", receipt.GasUsed)
		result["logs"] = logs
		result["logsBloom"] = hex.EncodeToString(bloom[:])
		result["type"] = fmt.Sprintf("0x%x", receipt.Type)
		result["status"] = fmt.Sprintf("0x%x", receipt.Status)

		if receipt.ContractAddress != (common.Address{}) {
			result["contractAddress"] = receipt.ContractAddress.String()
		}
		//log.Println(m)
	}
	return result, nil
}

func sendErrorMsg(method, code, msg string) error {
	requestBody := map[string]interface{}{}
	requestBody["method"] = method
	requestBody["code"] = code
	requestBody["message"] = msg

	pbytes, _ := json.Marshal(requestBody)
	httpRes, err := http.Post("https://account.meversemainnet.io/record-log", "application/json", bytes.NewBuffer([]byte(pbytes)))
	if err != nil {
		log.Println("err:", err)
		return err
	}

	defer httpRes.Body.Close()
	_, err = ioutil.ReadAll(httpRes.Body)
	if err != nil {
		log.Println("err:", err)
		return err
	}
	return nil
}

func appendLeftZeroPad(app []byte, size int, padd ...byte) []byte {
	if len(padd) < size {
		bs := make([]byte, size)
		copy(bs[size-len(padd):], padd)
		padd = bs
	}
	return append(app, padd...)
}

func (m *metamaskRelay) Version() uint16 {
	provider := m.cn.Provider()
	return provider.Version(provider.Height())
}

func (m *metamaskRelay) transactionHash(rlp string) (interface{}, error) {

	if len(rlp) <= 2 {
		return nil, errors.New("invalid tx")
	}

	rlpBytes, err := hex.DecodeString(strings.Replace(rlp, "0x", "", -1))
	if err != nil {
		return nil, err
	}
	etx := new(etypes.Transaction)
	if err := etx.UnmarshalBinary(rlpBytes); err != nil {
		return nil, err
	}

	var method string
	getSig := func() []byte {
		v, r, s := etx.RawSignatureValues()

		sig := []byte{}
		sig = appendLeftZeroPad(sig, 32, r.Bytes()...)
		sig = appendLeftZeroPad(sig, 32, s.Bytes()...)
		sig = append(sig, v.Bytes()...)

		return sig
	}

	// contract check
	ctx := m.cn.NewContext()
	gp := etx.GasPrice()
	if gp == nil || len(gp.Bytes()) == 0 {
		gp = etx.GasFeeCap()
	}

	to := common.Address{}
	if etx.To() != nil {
		to = *etx.To()
	}

	tx := &types.Transaction{
		ChainID:     m.cn.Provider().ChainID(),
		Version:     ctx.Version(ctx.TargetHeight()),
		Timestamp:   uint64(time.Now().UnixNano()),
		To:          to,
		Method:      method,
		Args:        rlpBytes,
		Seq:         etx.Nonce(),
		UseSeq:      true,
		IsEtherType: true,
		GasPrice:    gp,
		// VmType:      vmType,
	}

	err = m.nd.AddTx(tx, getSig())
	if err != nil {
		if strings.Contains(err.Error(), "future nonce") {
			return nil, err
		}
		var method string
		var code string
		if tx != nil {
			code = "100"
			method = tx.Method
		} else {
			code = "101"
			method = "emptytx"
		}
		bs, _err := json.Marshal(tx)
		if _err != nil {
			code = "102"
		}
		// sendErrorMsg(method, code, fmt.Sprintf("err:%+v from:%v tx:%v _err:%+v rlp:%v", err, From.String(), hex.EncodeToString(bs), _err, rlp))
		sendErrorMsg(method, code, fmt.Sprintf("err:%+v tx:%v _err:%+v rlp:%v", err, hex.EncodeToString(bs), _err, rlp))
		return nil, err
	}

	return tx.HashSig(), nil
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
	if tx.VmType != types.Evm {
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
			"hash":             tx.Hash(height).String(),
			"input":            input,
			"nonce":            nonce,
			"to":               to,
			"transactionIndex": fmt.Sprintf("0x%x", index),
			"value":            value,
			"v":                "0x" + hex.EncodeToString(sig[64:]),
			"r":                "0x" + hex.EncodeToString(sig[:32]),
			"s":                "0x" + hex.EncodeToString(sig[32:64]),
		}, nil
	} else {
		etx := new(etypes.Transaction)
		if err := etx.UnmarshalBinary(tx.Args); err != nil {
			return nil, err
		}

		var to interface{}
		if etx.To() != nil {
			to = etx.To().String()
		} else {
			to = nil
		}

		return map[string]interface{}{
			"blockHash":            bHash.String(),
			"blockNumber":          fmt.Sprintf("0x%x", height),
			"from":                 tx.From.String(),
			"gas":                  fmt.Sprintf("0x%x", etx.Gas()),
			"gasPrice":             fmt.Sprintf("0x%x", etx.GasPrice()),
			"maxFeePerGas":         fmt.Sprintf("0x%x", etx.GasFeeCap()),
			"maxPriorityFeePerGas": fmt.Sprintf("0x%x", etx.GasTipCap()),
			"hash":                 etx.Hash().String(),
			"input":                hex.EncodeToString(etx.Data()),
			"nonce":                etx.Nonce(),
			"to":                   to,
			"transactionIndex":     fmt.Sprintf("0x%x", index),
			"value":                fmt.Sprintf("0x%x", etx.Value()),
			"type":                 fmt.Sprintf("0x%x", etx.Type()),
			"accessList":           etx.AccessList(),
			"chainId":              fmt.Sprintf("0x%x", etx.ChainId()),
			"v":                    "0x" + hex.EncodeToString(sig[64:]),
			"r":                    "0x" + hex.EncodeToString(sig[:32]),
			"s":                    "0x" + hex.EncodeToString(sig[32:64]),
		}, nil
	}
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
			m := txparser.Abi(hex.EncodeToString(etx.Data()[:4]))
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
		txs = append(txs, tx.Hash(b.Header.Height).String())
	}

	bloom, err := bloomservice.BlockLogsBloom(m.cn, b)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"gasUsed":          "0x1c9c380",
		"gasLimit":         "0x1c9c380",
		"baseFeePerGas":    m.getBaseFeePerGas(),
		"hash":             bHash.String(),
		"logsBloom":        hex.EncodeToString(bloom[:]),
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
	ctx := m.cn.NewContext()
	{
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
		mabi := txparser.Abi(hex.EncodeToString(etx.Data()[:4]))
		if mabi.Name == "" {
			if !txparser.AbiCaches[etx.To().String()] {
				caller := viewchain.NewViewCaller(m.cn)
				txparser.AbiCaches[etx.To().String()] = true
				if rawabi, _, err := caller.Execute(*etx.To(), common.Address{}.String(), "abis", []interface{}{}); err == nil && len(rawabi) > 0 {
					if abis, ok := rawabi[0].([]interface{}); ok {
						strs := []string{}
						for _, funcStr := range abis {
							if f, ok := funcStr.(string); ok {
								strs = append(strs, f)
							}
						}
						bs := []byte("[" + strings.Join(strs, ",") + "]")
						reader := bytes.NewReader(bs)
						if abi, err := abi.JSON(reader); err == nil {
							for _, m := range abi.Methods {
								txparser.AddAbi(m)
							}
						}
					}
				}
				mabi = txparser.Abi(hex.EncodeToString(etx.Data()[:4]))
				if mabi.Name == "" {
					return nil, nil, errors.New("not exist abi")
				}
			} else {
				return nil, nil, errors.New("not exist abi")
			}
		}
		if isInvokeable {
			method = mabi.Name
		} else {
			method = strings.ToUpper(string(mabi.Name[0])) + mabi.Name[1:]
		}
	}
	if method == "" {
		m := txparser.Abi("69ca02dd")
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

func (m *metamaskRelay) ethCall(from, to, data string) (result string, gas uint64, err error) {
	if len(data) < 10 {
		//log.Println("ErrInvalidData len:", len(data))
		err = errors.WithStack(ErrInvalidData)
		return
	}
	toAddr, err := common.ParseAddress(to)
	if err != nil {
		return "", 0, err
	}
	if strings.Index(data, "0x") == 0 {
		data = data[2:]
	}

	ctx := m.cn.NewContext()
	if ctx.IsContract(toAddr) {
		caller := viewchain.NewViewCaller(m.cn)
		abiMs := txparser.Abis(toAddr, data[:8])
		// abiMs : map[string]abi.Method
		if len(abiMs) == 0 {
			if !txparser.AbiCaches[toAddr.String()] {
				txparser.AbiCaches[toAddr.String()] = true
				if rawabi, _, err := caller.Execute(toAddr, from, "abis", []interface{}{}); err == nil && len(rawabi) > 0 {
					if abis, ok := rawabi[0].([]interface{}); ok {
						strs := []string{}
						for _, funcStr := range abis {
							if f, ok := funcStr.(string); ok {
								strs = append(strs, f)
							}
						}
						bs := []byte("[" + strings.Join(strs, ",") + "]")
						reader := bytes.NewReader(bs)
						if abi, err := abi.JSON(reader); err == nil {
							for _, m := range abi.Methods {
								txparser.AddAbi(m)
							}
						}
					}
				}
				abiMs = txparser.Abis(toAddr, data[:8])
			}
		}
		if len(abiMs) == 0 {
			return "", 0, errors.New("func not found")
		}
		for _, abiM := range abiMs {
			if result, gas, err = execWithAbi(caller, from, toAddr, abiM, data); err == nil {
				return result, gas, nil
			}
		}
		return "", 0, err
	} else {
		// ethereum call

		fromAddr := common.HexToAddress(from)
		dataBytes, err := hex.DecodeString(data)
		if err != nil {
			return "", 0, err
		}
		result, err := m.DoCall(fromAddr, toAddr, dataBytes)
		if err != nil {
			return "", 0, err
		}
		if len(result.Revert()) > 0 {
			return "", 0, newRevertError(result)
		}
		return "0x" + hex.EncodeToString(result.ReturnData), result.UsedGas, result.Err
	}
}

func (m *metamaskRelay) DoCall(from, to common.Address, dataBytes []byte) (*core.ExecutionResult, error) {

	gas := uint64(math.MaxUint64 / 2)
	msg := etypes.NewMessage(from, &to, 0, big.NewInt(0), gas, big.NewInt(0), big.NewInt(0), big.NewInt(0), dataBytes, etypes.AccessList{}, true)

	ctx := m.cn.NewContext()
	statedb := types.NewStateDB(ctx)

	evm := defaultevm.DefaultEVM(statedb)

	// Execute the message.
	gp := new(ecore.GasPool).AddGas(math.MaxUint64)
	result, err := mcore.ApplyMessage(evm, msg, gp)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (m *metamaskRelay) getBaseFeePerGas() string {
	return "0x23400641"
}

func execWithAbi(caller *viewchain.ViewCaller, from string, toAddr ecommon.Address, abiM abi.Method, data string) (string, uint64, error) {
	bs, err := hex.DecodeString(data)
	if err != nil {
		return "", 0, err
	}
	obj, err := txparser.Inputs(bs)
	if err != nil {
		return "", 0, err
	}
	output, gas, err := caller.Execute(toAddr, from, abiM.Name, obj)
	if err != nil {
		err = fmt.Errorf("%v call %v method %v", err.Error(), toAddr, abiM.Name)
		return "", 0, err
	}

	bs, err = txparser.Outputs(abiM.Sig, output)
	if err != nil {
		return "", 0, err
	}
	return "0x" + hex.EncodeToString(bs), gas, nil
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
