package txparser

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

var (
	funcSigs          map[string]map[string]abi.Method
	AbiCaches         map[string]bool
	SUPPORTSINTERFACE string = ERCFuncSignature("supportsInterface(bytes4)")
	NAME              string = ERCFuncSignature("name()")
	SYMBOL            string = ERCFuncSignature("symbol()")
	DECIMALS          string = ERCFuncSignature("decimals()")
	TOTALSUPPLY       string = ERCFuncSignature("totalSupply()")
	BALANCEOF         string = ERCFuncSignature("balanceOf(address)")
)

/*
--ERC20
allowance(address,address)
approve(address,uint256)
transferFrom(address,address,uint256)
totalSupply()
balanceOf(address)
transfer(address,uint256)

--ERC721
approve(address,uint256)
ownerOf(uint256)
safeTransferFrom(address,address,uint256)
safeTransferFrom(address,address,uint256,bytes)
supportsInterface(bytes4)
balanceOf(address)
getApproved(uint256)
isApprovedForAll(address,address)
setApprovalForAll(address,bool)
transferFrom(address,address,uint256)

--ERC1155
safeTransferFrom(address,address,uint256,uint256,bytes)
setApprovalForAll(address,bool)
supportsInterface(bytes4)
balanceOf(address,uint256)
balanceOfBatch(address[],uint256[])
isApprovedForAll(address,address)
safeBatchTransferFrom(address,address,uint256[],uint256[],bytes)
*/

func init() {
	funcSigs = map[string]map[string]abi.Method{}
	AbiCaches = map[string]bool{}
	ercs := [][]byte{
		IERC20,
		IERC721,
		IERC1155,
		Token,
		Formulator,
		Factory,
		Router,
		Bridge,
		SwapUni,
		SwapCurve,
		Farm,
		Pool,
		Imo,
		DepositUSDT,
		NFT721,
		MarketDATA,
		MarketOP,
		Reveal,
		Request,
	}
	for _, v := range ercs {
		reader := bytes.NewReader(v)
		a, err := abi.JSON(reader)
		if err != nil {
			panic(err)
		}
		for _, m := range a.Methods {
			AddAbi(m)
		}
	}
}

func AddAbi(m abi.Method) {
	outputArr := make([]string, len(m.Outputs))
	for i, o := range m.Outputs {
		outputArr[i] = o.Type.String()
	}
	output := strings.Join(outputArr, ",")

	outMap := funcSigs[m.Sig]
	if outMap == nil {
		outMap = map[string]abi.Method{}
		funcSigs[m.Sig] = outMap
		funcSigs[hex.EncodeToString(m.ID)] = outMap
	}
	outMap[output] = m

	// name = name + ""
	// log.Println(name, m.Name, m.RawName, m.Sig, hex.EncodeToString(m.ID))
}

func Abi(method string) (abi.Method, error) {
	ms := funcSigs[method]
	var m abi.Method
	for _, m = range ms {
		break
	}
	return m, nil
}

func Abis(cont common.Address, method string) map[string]abi.Method {
	return funcSigs[method]
}

func Inputs(data []byte) ([]interface{}, error) {
	if len(data) < 4 {
		return nil, errors.New("not found func sig")
	}
	method := hex.EncodeToString(data[:4])

	ms := funcSigs[method]
	var m abi.Method
	for _, m = range ms {
		break
	}
	obj, err := m.Inputs.Unpack(data[4:])
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func Outputs(method string, data []interface{}) ([]byte, error) {
	ms := funcSigs[method]
	for i, v := range data {
		if a, ok := v.(*amount.Amount); ok {
			data[i] = a.Int
		} else if as, ok := v.([]*amount.Amount); ok {
			bis := []*big.Int{}
			for _, a := range as {
				bis = append(bis, a.Int)
			}
			data[i] = bis
		}
		// switch tv := v.(type) {
		// case int:
		// 	data[i] = big.NewInt(0).SetInt64(int64(tv))
		// case int8:
		// 	data[i] = big.NewInt(0).SetInt64(int64(tv))
		// case int16:
		// 	data[i] = big.NewInt(0).SetInt64(int64(tv))
		// case int32:
		// 	data[i] = big.NewInt(0).SetInt64(int64(tv))
		// case int64:
		// 	data[i] = big.NewInt(0).SetInt64(int64(tv))
		// case uint:
		// 	data[i] = big.NewInt(0).SetUint64(uint64(tv))
		// case uint8:
		// 	data[i] = big.NewInt(0).SetUint64(uint64(tv))
		// case uint16:
		// 	data[i] = big.NewInt(0).SetUint64(uint64(tv))
		// case uint32:
		// 	data[i] = big.NewInt(0).SetUint64(uint64(tv))
		// case uint64:
		// 	data[i] = big.NewInt(0).SetUint64(uint64(tv))
		// }
	}

	var err error
	var bs []byte
	for _, m := range ms {
		if len(m.Outputs) != len(data) {
			err = errors.New("invalid output count")
			continue
		}
		for i, ot := range m.Outputs {
			if ot.Type.GetType() == reflect.TypeOf(data[i]) {
				continue
			}
			switch ot.Type.T {
			case abi.IntTy:
				data[i], err = reflectIntType(false, ot.Type.Size, data[i])
			case abi.UintTy:
				data[i], err = reflectIntType(true, ot.Type.Size, data[i])
			case abi.BoolTy:
				sr, ok := data[i].(fmt.Stringer)
				var s string
				if ok {
					s = sr.String()
				} else {
					s, _ = data[i].(string)
				}
				if s == "true" {
					data[i] = true
				} else if s == "false" {
					data[i] = false
				}
			case abi.StringTy:
				switch t := data[i].(type) {
				case string:
					data[i] = t
				case []byte:
					data[i] = string(t)
				case fmt.Stringer:
					data[i] = t.String()
				}
			case abi.SliceTy, abi.ArrayTy:
				if s, ok := data[i].([]interface{}); ok {
					sts := []string{}
					for _, v := range s {
						sts = append(sts, fmt.Sprintf("%v", v))
					}
					data[i] = sts
				}
			case abi.AddressTy:
				if s, ok := data[i].(string); ok {
					data[i] = common.HexToAddress(s)
				}
			case abi.BytesTy:
				if s, ok := data[i].(string); ok {
					data[i], err = hex.DecodeString(s)
				}
			case abi.HashTy:
				if s, ok := data[i].(string); ok {
					data[i] = hash.HexToHash(s)
				}
			// case abi.FixedPointTy:
			// 	// fixedpoint type currently not used
			// 	return reflect.ArrayOf(32, reflect.TypeOf(byte(0)))
			// case abi.FunctionTy:
			// 	return reflect.ArrayOf(24, reflect.TypeOf(byte(0)))
			default:
				panic("Invalid type")
			}
			if err != nil {
				break
			}
		}
		if err != nil {
			continue
		}
		bs, err = m.Outputs.Pack(data...)
		if err == nil {
			break
		}
	}
	return bs, err
}

func reflectIntType(unsigned bool, size int, data interface{}) (i interface{}, err error) {
	s := fmt.Sprintf("%v", data)
	base := 10
	if strings.Contains(s, "0x") {
		base = 16
		s = strings.Replace(s, "0x", "", -1)
	}

	if unsigned {
		var ui uint64
		switch size {
		case 8:
			ui, err = strconv.ParseUint(s, base, 8)
			i = big.NewInt(0).SetUint64(uint64(ui))
		case 16:
			ui, err = strconv.ParseUint(s, base, 16)
			i = big.NewInt(0).SetUint64(uint64(ui))
		case 32:
			ui, err = strconv.ParseUint(s, base, 32)
			i = big.NewInt(0).SetUint64(uint64(ui))
		case 64:
			ui, err = strconv.ParseUint(s, base, 64)
			i = big.NewInt(0).SetUint64(uint64(ui))
		case 256:
			if bi, ok := big.NewInt(0).SetString(s, base); ok {
				i = bi
			} else {
				err = errors.New("invalid bigInt value")
			}
		}
		return
	}
	var si int64
	switch size {
	case 8:
		si, err = strconv.ParseInt(s, base, 8)
		i = big.NewInt(0).SetInt64(int64(si))
	case 16:
		si, err = strconv.ParseInt(s, base, 16)
		i = big.NewInt(0).SetInt64(int64(si))
	case 32:
		si, err = strconv.ParseInt(s, base, 32)
		i = big.NewInt(0).SetInt64(int64(si))
	case 64:
		si, err = strconv.ParseInt(s, base, 64)
		i = big.NewInt(0).SetInt64(int64(si))
	case 256:
		if bi, ok := big.NewInt(0).SetString(s, base); ok {
			i = bi
		} else {
			err = errors.New("invalid bigInt value")
		}
	}
	return
}

func ERCFuncSignature(fn string) string { // ex"transfer(address,uint256)"
	transferFnSignature := []byte(fn)
	hash := sha3.NewLegacyKeccak256()
	hash.Write(transferFnSignature)
	return hex.EncodeToString(hash.Sum(nil)[:4])
}

func EthTxFromRLP(rlpBytes []byte) (*etypes.Transaction, []byte, error) {
	if len(rlpBytes) == 0 {
		return nil, nil, errors.New("invalid tx")
	}

	var t etypes.Transaction
	etx, err := &t, t.UnmarshalBinary(rlpBytes)
	if err != nil {
		return nil, nil, err
	}
	v, r, s := etx.RawSignatureValues()

	sig := []byte{}
	sig = appendLeftZeroPad(sig, 32, r.Bytes()...)
	sig = appendLeftZeroPad(sig, 32, s.Bytes()...)
	sig = append(sig, v.Bytes()...)

	return etx, sig, nil
}

func appendLeftZeroPad(app []byte, size int, padd ...byte) []byte {
	if len(padd) < size {
		bs := make([]byte, size)
		copy(bs[size-len(padd):], padd)
		padd = bs
	}
	return append(app, padd...)
}
