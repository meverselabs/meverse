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
	"runtime"
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
	"github.com/meverselabs/meverse/core/ctypes"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/ethereum/core"
	mcore "github.com/meverselabs/meverse/ethereum/core"
	"github.com/meverselabs/meverse/ethereum/core/defaultevm"
	mtypes "github.com/meverselabs/meverse/ethereum/core/types"
	"github.com/meverselabs/meverse/ethereum/core/vm"
	"github.com/meverselabs/meverse/ethereum/eth/tracers"
	"github.com/meverselabs/meverse/ethereum/eth/tracers/logger"
	_ "github.com/meverselabs/meverse/ethereum/eth/tracers/native"
	"github.com/meverselabs/meverse/ethereum/ethapi"
	"github.com/meverselabs/meverse/ethereum/params"
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

	idx int
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
		return fmt.Sprintf("%v", m.chainID.String()), nil
	})
	s.Set("eth_symbol", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		addr := cn.Store().MainToken()
		ctx := cn.NewContext()
		cont, err := ctx.Contract(*addr)
		if err != nil {
			return nil, err
		}
		if tcont, ok := cont.(*token.TokenContract); ok {
			bcc := ctx.ContractContext(cont, common.ZeroAddr)
			return tcont.Symbol(bcc), nil
		}
		return "MEV", nil
	})
	s.Set("eth_feeHistory", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		oldestBlock, _ := arg.String(1)
		if oldestBlock == "latest" || oldestBlock == "pending" {
			cheight := cn.Provider().Height()
			oldestBlock = strconv.FormatUint(uint64(cheight), 16)
		}
		gasPrice := fmt.Sprintf("0x%x", m.basicFee())
		return map[string]interface{}{
			"oldestBlock":   oldestBlock,
			"baseFeePerGas": []string{gasPrice, gasPrice},
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
		if hei < uint64(cn.Provider().InitHeight()) {
			return nil, errors.New("not found block")
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
			return fmt.Sprintf("0x%x", m.basicFee()), nil
		}
	})

	s.Set("eth_estimateGas", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		switch m.Version() {
		case 0, 1:
			return "0x0186A0", nil
			//return "0x1DCD6500", nil
		default:
			argMap, err := arg.Map(0)
			if err != nil {
				return nil, err
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

			toAddr, err := common.ParseAddress(to)
			if err != nil {
				return nil, err
			}

			ctx := m.cn.NewContext()
			if !ctx.IsContract(toAddr) {

				fromAddr := common.HexToAddress(from)

				// error when argMap.value > balance
				value, err := getValueFromParam(argMap["value"])
				if err != nil {
					return nil, err
				} else {
					if value.Cmp(new(big.Int)) != 0 {
						statedb := types.NewStateDB(ctx)
						balance := statedb.GetBalance(fromAddr) // from can't be nil
						if value.Cmp(balance) >= 0 {
							return nil, errors.New("insufficient funds for transfer")
						}
					}
				}

				// ethereum 의 경우만 적용
				var (
					lo  uint64 = params.TxGas
					hi  uint64 = params.BlockGasLimit
					cap uint64 = hi
				)

				if strings.Index(data, "0x") == 0 {
					data = data[2:]
				}
				dataBytes, err := hex.DecodeString(data)
				if err != nil {
					return nil, err
				}

				// Create a helper to check if a gas allowance results in an executable transaction
				executable := func(gas uint64) (bool, *core.ExecutionResult, error) {
					result, err := m.DoCall(fromAddr, toAddr, dataBytes, gas, value)
					if err != nil {
						if errors.Is(err, core.ErrIntrinsicGas) {
							return true, nil, nil // Special case, raise gas limit
						}
						return true, nil, err // Bail out
					}
					return result.Failed(), result, nil
				}

				for lo+1 < hi {
					mid := (hi + lo) / 2
					failed, _, err := executable(mid)
					// If the error is not nil(consensus error), it means the provided message
					// call or transaction will never be accepted no matter how much gas it is
					// assigned. Return the error directly, don't struggle any more.
					if err != nil {
						return nil, err
					}
					if failed {
						lo = mid
					} else {
						hi = mid
					}
				}

				// Reject the transaction as invalid if it still fails at the highest allowance
				if hi == cap {
					failed, result, err := executable(hi)
					if err != nil {
						return nil, err
					}
					if failed {
						if result != nil && result.Err != vm.ErrOutOfGas {
							if len(result.Revert()) > 0 {
								return 0, apiserver.NewRevertError(result)
							}
							return nil, result.Err
						}
						// Otherwise, the specified gas cap is too low
						return nil, fmt.Errorf("gas required exceeds allowance (%d)", cap)
					}

				}
				return fmt.Sprintf("0x%x", hi), nil
			} else {
				// ethereum 제외
				_, gas, err := m.ethCall(from, to, data, 0, new(big.Int), false)
				return fmt.Sprintf("0x%x", gas), err
			}
		}
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

		value, err := getValueFromParam(param["value"])
		if err != nil {
			return nil, err
		}

		result, _, err := m.ethCall(from, to, data, 0, value, true)

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

	s.Set("debug_traceCall", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return m.traceCall(ID, arg)
	})

	s.Set("debug_traceTransaction", func(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
		return m.traceTx(ID, arg)
	})
}

// get value from json-rpc
func getValueFromParam(valueParam interface{}) (*big.Int, error) {
	var value *big.Int
	if valueParam != nil {
		if valueStr, ok := valueParam.(string); ok {
			if strings.Index(valueStr, "0x") == 0 {
				valueStr = valueStr[2:]
				var ok2 bool
				value, ok2 = new(big.Int).SetString(valueStr, 16)
				if !ok2 {
					return nil, errors.New("invalid parameter value")
				}
			} else {
				var ok2 bool
				value, ok2 = new(big.Int).SetString(valueStr, 10)
				if !ok2 {
					return nil, errors.New("invalid parameter value")
				}
			}
		} else {
			return nil, errors.New("invalid parameter value")
		}
	} else {
		value = new(big.Int)
	}
	return value, nil
}

// cumulativeGasUsed return block's cumulativeGasUsed  up to idx (idx not included)
func (m *metamaskRelay) blockTotalGas(b *types.Block) (*big.Int, error) {
	return m.cumulativeGasUsed(b, uint16(len(b.Body.Transactions)))
}

// cumulativeGasUsed return block's cumulativeGasUsed  up to idx (idx not included)
func (m *metamaskRelay) cumulativeGasUsed(b *types.Block, idx uint16) (*big.Int, error) {

	size := uint16(len(b.Body.Transactions))
	if idx > size {
		return nil, errors.New("tranaction index is out of range")
	}
	gasPrice := m.basicFee()
	totalGas := new(big.Int)
	var receipts types.Receipts
	for i := uint16(0); i < idx; i++ {
		tx := b.Body.Transactions[i]
		if tx.VmType != types.Evm {
			fee, err := findTxFeeFromEvent(b.Body.Events, i)
			if err != nil {
				return nil, err
			}
			totalGas = new(big.Int).Add(totalGas, fee.Int.Div(fee.Int, gasPrice.Int))
		} else {
			if receipts == nil {
				var err error
				receipts, err = m.cn.Provider().Receipts(b.Header.Height)
				if err != nil {
					return nil, err
				}
				if int(idx) > len(receipts) {
					return nil, errors.New("tranaction index is out of range")
				}
			}

			receipt := receipts[i]
			totalGas = new(big.Int).Add(totalGas, new(big.Int).SetUint64(receipt.CumulativeGasUsed))
		}
	}
	return totalGas, nil
}

func getReceipt(tx *types.Transaction, b *types.Block, TxID itxsearch.TxID, m *metamaskRelay, bHash ecommon.Hash) (map[string]interface{}, error) {

	gasPrice := m.basicFee()

	cumulativeGasUsed, err := m.cumulativeGasUsed(b, TxID.Index)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	var receipt *etypes.Receipt

	if tx.VmType != types.Evm {

		fee, err := findTxFeeFromEvent(b.Body.Events, TxID.Index)
		if err != nil {
			return nil, err
		}
		// admin tranaction fee = nil
		if fee == nil {
			fee = amount.NewAmount(0, 0)
		}

		// gasUsed = fee / EffectiveGasPrice
		gasUsed := fee.Int.Div(fee.Int, gasPrice.Int)
		cumulativeGasUsed = new(big.Int).Add(cumulativeGasUsed, new(big.Int).SetUint64(gasUsed.Uint64()))

		result = map[string]interface{}{
			"blockHash":         bHash.String(),
			"blockNumber":       fmt.Sprintf("0x%x", TxID.Height),
			"transactionHash":   tx.Hash(TxID.Height),
			"transactionIndex":  fmt.Sprintf("0x%x", TxID.Index),
			"from":              tx.From.String(),
			"to":                tx.To.String(),
			"gasUsed":           fmt.Sprintf("0x%x", gasUsed),
			"cumulativeGasUsed": fmt.Sprintf("0x%x", cumulativeGasUsed),
			"effectiveGasPrice": fmt.Sprintf("0x%x", gasPrice),
			"contractAddress":   nil, // or null, if none was created
			"type":              "0x2",
			"status":            "0x1", //TODO 성공 1 실패 0
			"data":              map[string]string{},
		}
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
		receipt = receipts[TxID.Index]
		signer := mtypes.MakeSigner(m.cn.Provider().ChainID(), TxID.Height)
		if err := receipts.DeriveReceiptFields(bHash, uint64(TxID.Height), TxID.Index, etx, signer); err != nil {
			return nil, err
		}

		var to interface{}
		if etx.To() != nil {
			to = etx.To().String()
		} else {
			to = nil
		}

		cumulativeGasUsed = new(big.Int).Add(cumulativeGasUsed, new(big.Int).SetUint64(receipt.CumulativeGasUsed))

		result = map[string]interface{}{
			"blockHash":         bHash.String(),
			"blockNumber":       fmt.Sprintf("0x%x", TxID.Height),
			"transactionHash":   eHash.String(),
			"transactionIndex":  fmt.Sprintf("0x%x", TxID.Index),
			"from":              tx.From.String(),
			"to":                to,
			"cumulativeGasUsed": fmt.Sprintf("0x%x", cumulativeGasUsed),
			"gasUsed":           fmt.Sprintf("0x%x", receipt.CumulativeGasUsed),
			"effectiveGasPrice": fmt.Sprintf("0x%x", gasPrice),
			"type":              fmt.Sprintf("0x%x", receipt.Type),
			"status":            fmt.Sprintf("0x%x", receipt.Status),
			"data":              map[string]string{},
		}

		if receipt.ContractAddress != (common.Address{}) {
			result["contractAddress"] = receipt.ContractAddress.String()
		}
		//log.Println(m)
	}

	bloom, logs, err := bloomservice.TxLogsBloom(m.cn, b, TxID.Index, receipt)
	if err != nil {
		return nil, err
	}

	result["logs"] = logs
	result["logsBloom"] = hex.EncodeToString(bloom[:])

	return result, nil
}

// findTxFeeFromEvent find tx's fee(gasUsed) from given event list
func findTxFeeFromEvent(evs []*ctypes.Event, idx uint16) (*amount.Amount, error) {

	if len(evs) == 0 {
		return amount.NewAmount(0, 0), nil
	}
	for _, ev := range evs {
		if ev.Type != ctypes.EventTagTxFee {
			continue
		}
		if ev.Index == idx {
			is, err := bin.TypeReadAll(ev.Result, 1)
			if err != nil {
				return nil, err
			}
			return is[0].(*amount.Amount), nil
		}
	}

	// admin tranaction (ex. Contract.Deploy) fee = nil
	return amount.NewAmount(0, 0), nil
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

	// contract check
	ctx := m.cn.NewContext()
	gp := etx.GasPrice()

	// if gp == nil || gp.Cmp(m.basicFee().Int) != 0 {
	// 	return nil, ErrInvalidGasPrice
	// }

	if gp == nil || len(gp.Bytes()) == 0 {
		gp = etx.GasFeeCap()
	}

	var method string
	// var vmType uint8

	if len(etx.Data()) > 4 {
		method = "0x" + hex.EncodeToString(etx.Data()[:4])
	}

	// getMethod := func() (string, error) {
	// 	m := txparser.Abi(hex.EncodeToString(etx.Data()[:4]))
	// 	if m.Name == "" {
	// 		return "", errors.New("not exist abi")
	// 	}
	// 	return m.Name, nil
	// }

	getSig := func() []byte {
		v, r, s := etx.RawSignatureValues()

		sig := []byte{}
		sig = appendLeftZeroPad(sig, 32, r.Bytes()...)
		sig = appendLeftZeroPad(sig, 32, s.Bytes()...)
		sig = append(sig, v.Bytes()...)

		return sig
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

	// fmt.Printf("tx = %+v\n", tx)
	// fmt.Printf("rlp = %+v\n", rlp)

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
	// if vmType != types.Evm {
	// 	return tx.HashSig(), nil
	// }
	// return etx.Hash(), nil
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

	return getTransactionMap(m, &bHash, height, index, tx, sig)
}

func getTransactionMap(m *metamaskRelay, bHash *ecommon.Hash, height uint32, index uint16, tx *types.Transaction, sig common.Signature) (interface{}, error) {

	gasPrice := m.basicFee()

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
		vbs := sig[64:]
		if len(vbs) == 0 {
			vbs = []byte{0}
		}

		result := &ethapi.RPCTransaction{
			BlockHash:        bHash.String(),
			BlockNumber:      fmt.Sprintf("0x%x", height),
			From:             tx.From.String(),
			Gas:              "0x1f4b698",
			GasPrice:         fmt.Sprintf("0x%x", gasPrice), // "0x71e0e496c",
			Hash:             tx.Hash(height).String(),
			Input:            input,
			Nonce:            nonce,
			To:               to,
			TransactionIndex: fmt.Sprintf("0x%x", index),
			Value:            value,
			ChainID:          fmt.Sprintf("0x%x", m.cn.Provider().ChainID()),
			V:                "0x" + hex.EncodeToString(vbs),
			R:                "0x" + hex.EncodeToString(sig[:32]),
			S:                "0x" + hex.EncodeToString(sig[32:64]),
		}

		return result, nil

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
		vbs := sig[64:]
		if len(vbs) == 0 {
			vbs = []byte{0}
		}

		result := &ethapi.RPCTransaction{
			BlockHash:        bHash.String(),
			BlockNumber:      fmt.Sprintf("0x%x", height),
			From:             tx.From.String(),
			Gas:              fmt.Sprintf("0x%x", etx.Gas()),
			GasPrice:         fmt.Sprintf("0x%x", gasPrice), // fmt.Sprintf("0x%x", etx.GasPrice()),
			GasFeeCap:        fmt.Sprintf("0x%x", etx.GasFeeCap()),
			GasTipCap:        fmt.Sprintf("0x%x", etx.GasTipCap()),
			Hash:             etx.Hash().String(),
			Input:            hex.EncodeToString(etx.Data()),
			Nonce:            fmt.Sprintf("0x%x", etx.Nonce()),
			To:               to,
			TransactionIndex: fmt.Sprintf("0x%x", index),
			Value:            fmt.Sprintf("0x%x", etx.Value()),
			Type:             fmt.Sprintf("0x%x", etx.Type()),
			ChainID:          fmt.Sprintf("0x%x", etx.ChainId()),
			V:                "0x" + hex.EncodeToString(vbs),
			R:                "0x" + hex.EncodeToString(sig[:32]),
			S:                "0x" + hex.EncodeToString(sig[32:64]),
		}

		return result, nil
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

	txs := []interface{}{}
	for idx, tx := range b.Body.Transactions {
		if fullTx {
			sig := b.Body.TransactionSignatures[idx]
			result, err := getTransactionMap(m, &bHash, uint32(hei), uint16(idx), tx, sig)
			if err != nil {
				return nil, err
			}
			txs = append(txs, result)
		} else {
			txs = append(txs, tx.Hash(b.Header.Height).String())
		}
	}

	bloom, err := bloomservice.BlockLogsBloom(m.cn, b)
	if err != nil {
		return nil, err
	}

	totalGasUsed, err := m.blockTotalGas(b)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"gasUsed": fmt.Sprintf("0x%x", totalGasUsed),
		//"gasLimit":         fmt.Sprintf("0x%x", params.BlockGasLimit),
		//"baseFeePerGas":    fmt.Sprintf("0x%x", gasPrice),
		"hash":             bHash.String(),
		"logsBloom":        "0x" + hex.EncodeToString(bloom[:]),
		"number":           fmt.Sprintf("0x%x", b.Header.Height),
		"parentHash":       b.Header.PrevHash.String(),
		"size":             fmt.Sprintf("0x%x", len(b.Body.Transactions)),
		"timestamp":        fmt.Sprintf("0x%x", b.Header.Timestamp/1000),
		"transactions":     txs,
		"transactionsRoot": b.Header.LevelRootHash.String(),
		"extraData":        "0x00",
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
		funcSig := hex.EncodeToString(etx.Data()[:4])
		mabi := txparser.Abi(funcSig)
		if mabi.Name == "" {
			caller := viewchain.NewViewCaller(m.cn)
			mabi, err = txparser.MakeAbi(*etx.To(), funcSig, caller)
			if err != nil {
				return nil, nil, err
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

	// if isInvokeable {
	// 	data, _, err := types.TxArg(ctx, tx)
	// 	if err != nil {
	// 		return nil, nil, err
	// 	}
	// 	tx.IsEtherType = false
	// 	tx.Args = append(bin.TypeWriteAll(data), tx.Args...)
	// }

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
	rv = strings.Replace(rv, "0x0", "0x", 1)
	if rv == "0x" {
		rv = "0x0"
	}
	return rv, nil
}

func (m *metamaskRelay) basicFee() *amount.Amount {
	return m.cn.NewContext().BasicFee()
}

func (m *metamaskRelay) ethCall(from, to, data string, inputGas uint64, value *big.Int, needResult bool) (result string, gas uint64, err error) {
	// if len(data) < 10 {
	// 	//log.Println("ErrInvalidData len:", len(data))
	// 	err = errors.WithStack(ErrInvalidData)
	// 	return
	// }
	toAddr, err := common.ParseAddress(to)
	if err != nil {
		return "", 0, err
	}
	if strings.Index(data, "0x") == 0 {
		data = data[2:]
	}
	if len(data) < 8 {
		return "", 0, errors.New("invalid data size")
	}

	ctx := m.cn.NewContext()
	if ctx.IsContract(toAddr) {
		caller := viewchain.NewViewCaller(m.cn)
		abiMs, err := txparser.Abis(toAddr, data[:8], caller)
		if err != nil {
			return "", 0, err
		}
		for _, abiM := range abiMs {
			if needResult {
				if output, gas, err := execWithAbi(caller, from, toAddr, abiM, data); err == nil {
					result, err = getOutput(abiM, output)
					return result, gas, err
				}
			} else {
				if _, gas, err = execWithAbi(caller, from, toAddr, abiM, data); err == nil {
					return result, gas, nil
				}
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
		if inputGas == 0 {
			inputGas = uint64(math.MaxUint64 / 2)
		}
		result, err := m.DoCall(fromAddr, toAddr, dataBytes, inputGas, value)
		if err != nil {
			return "", 0, err
		}
		if len(result.Revert()) > 0 {
			return "", 0, apiserver.NewRevertError(result)
		}
		if result.Err != nil {
			return "0x" + hex.EncodeToString(result.ReturnData), result.UsedGas, result.Err
		}
		return "0x" + hex.EncodeToString(result.ReturnData), result.UsedGas, result.Err
	}
}

func (m *metamaskRelay) DoCall(from, to common.Address, dataBytes []byte, gas uint64, value *big.Int) (res *core.ExecutionResult, err error) {
	defer func() {
		r := recover()
		if _, ok := r.(runtime.Error); ok {
			msg := fmt.Sprintf("%v", r)
			res = nil
			err = errors.New(msg)
		}
	}()

	//gas := uint64(math.MaxUint64 / 2)
	msg := etypes.NewMessage(from, &to, 0, value, gas, big.NewInt(0), big.NewInt(0), big.NewInt(0), dataBytes, etypes.AccessList{}, true)

	ctx := m.cn.NewContext()
	statedb := types.NewStateDB(ctx)

	evm := defaultevm.DefaultEVM(statedb, nil)

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

// traceCall returns  Tracer from json-rpc call (dubug_traceCall, debug_traceTransaction)
func tracer(arg *apiserver.Argument, idx int) (tracers.Tracer, error) {

	config := &tracers.TraceConfig{}
	if configMap, _ := arg.Map(idx); configMap != nil {
		if jsonStr, err := json.Marshal(configMap); err != nil {
			return nil, err
		} else {
			if err := json.Unmarshal(jsonStr, config); err != nil {
				return nil, err
			}
		}
	}
	var tracer tracers.Tracer
	tracer = logger.NewStructLogger(&config.Config)
	if config.Tracer != "" {
		var err error
		tracer, err = tracers.DefaultDirectory.New(config.Tracer, new(tracers.Context), config.TracerConfig)
		if err != nil {
			return nil, err
		}
	}

	return tracer, nil
}

// traceCall excutes debug_traceCall json-rpc call
func (m *metamaskRelay) traceCall(ID interface{}, arg *apiserver.Argument) (interface{}, error) {
	// var tracer tracers.Tracer
	//var config *tracers.TraceConfig
	var from, to common.Address
	var data []byte
	var err error

	param, _ := arg.Map(0)
	if tol, ok := param["to"].(string); ok {
		to, err = common.ParseAddress(tol)
		if err != nil {
			return nil, err
		}
	}

	ctx := m.cn.NewContext()

	// only evm
	if ctx.IsContract(to) {
		return nil, errors.New("only for evm contract")
	}

	// only latest
	if block, err := arg.String(1); err != nil {
		return nil, err
	} else if block != "latest" {
		return nil, errors.New("only latest is allowd")
	}

	if froml, ok := param["from"].(string); ok {
		from, err = common.ParseAddress(froml)
		if err != nil {
			return nil, err
		}
	}

	if datal, ok := param["data"].(string); ok {
		if strings.HasPrefix(datal, "0x") {
			datal = datal[2:]
		}
		data, err = hex.DecodeString(datal)
		if err != nil {
			return nil, err
		}
	}

	value, err := getValueFromParam(param["value"])
	if err != nil {
		return nil, err
	}

	msg := etypes.NewMessage(from, &to, 0, value, uint64(math.MaxUint64/2), big.NewInt(0), big.NewInt(0), big.NewInt(0), data, etypes.AccessList{}, true)

	// config
	tracer, err := tracer(arg, 2)
	if err != nil {
		return nil, err
	}

	evm := defaultevm.DefaultEVM(types.NewStateDB(ctx), tracer)
	_, err = mcore.ApplyMessage(evm, msg, new(ecore.GasPool).AddGas(math.MaxUint64))
	if err != nil {
		return nil, err
	}

	return tracer.GetResult()
}

// traceTx excutes debug_traceTransaction json-rpc call
func (m *metamaskRelay) traceTx(ID interface{}, arg *apiserver.Argument) (interface{}, error) {

	var tx *types.Transaction

	if txhash, err := arg.String(0); err != nil {
		return nil, err
	} else {
		hs := hash.HexToHash(txhash)
		TxID, err := m.ts.TxIndex(hs)
		if err != nil {
			return nil, err
		}
		if TxID.Err != nil {
			return nil, TxID.Err
		}

		b, err := m.cn.Provider().Block(TxID.Height)
		if err != nil {
			return nil, err
		}

		if int(TxID.Index) >= len(b.Body.Transactions) {
			return nil, errors.New("invalid txhash")
		}
		tx = b.Body.Transactions[TxID.Index]
	}

	// only evm
	if tx.VmType != types.Evm {
		return nil, errors.New("only for evm contract")
	}
	etx := new(etypes.Transaction)
	if err := etx.UnmarshalBinary(tx.Args); err != nil {
		return nil, err
	}

	msg := etypes.NewMessage(tx.From, etx.To(), 0, etx.Value(), uint64(math.MaxUint64/2), big.NewInt(0), big.NewInt(0), big.NewInt(0), etx.Data(), etx.AccessList(), true)

	// tracer
	tracer, err := tracer(arg, 1)
	if err != nil {
		return nil, err
	}

	// execute tx
	evm := defaultevm.DefaultEVM(types.NewStateDB(m.cn.NewContext()), tracer)
	_, err = mcore.ApplyMessage(evm, msg, new(ecore.GasPool).AddGas(math.MaxUint64))
	if err != nil {
		return nil, err
	}

	return tracer.GetResult()
}

// func execWithAbi(caller *viewchain.ViewCaller, from string, toAddr ecommon.Address, abiM abi.Method, data string) (string, uint64, error) {
// 	output, gas, err := newFunction(caller, from, toAddr, abiM, data)
// 	if err != nil {
// 		return "", 0, err
// 	}

// 	return getOutput(abiM, output, gas)
// }

func execWithAbi(caller *viewchain.ViewCaller, from string, toAddr ecommon.Address, abiM abi.Method, data string) ([]interface{}, uint64, error) {
	bs, err := hex.DecodeString(data)
	if err != nil {
		return nil, 0, err
	}
	obj, err := txparser.Inputs(bs)
	if err != nil {
		return nil, 0, err
	}
	output, gas, err := caller.Execute(toAddr, from, abiM.Name, obj)
	if err != nil {
		err = fmt.Errorf("%v call %v method %v", err.Error(), toAddr, abiM.Name)
		return nil, 0, err
	}
	return output, gas, nil
}

func getOutput(abiM abi.Method, output []interface{}) (string, error) {
	bs, err := txparser.Outputs(abiM.Sig, output)
	if err != nil {
		return "", err
	}
	return "0x" + hex.EncodeToString(bs), nil
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
