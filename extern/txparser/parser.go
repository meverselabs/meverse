package txparser

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

var (
	FuncSigs          map[string]map[string]abi.Method
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
	FuncSigs = map[string]map[string]abi.Method{}
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
	}
	for _, v := range ercs {
		reader := bytes.NewReader(v)
		a, err := abi.JSON(reader)
		if err != nil {
			panic(err)
		}
		for name, m := range a.Methods {
			outputArr := make([]string, len(m.Outputs))
			for i, o := range m.Outputs {
				outputArr[i] = o.Type.String()
			}
			output := strings.Join(outputArr, ",")

			outMap := FuncSigs[m.Sig]
			if outMap == nil {
				outMap = map[string]abi.Method{}
			}
			outMap[output] = m

			FuncSigs[m.Sig] = outMap
			FuncSigs[hex.EncodeToString(m.ID)] = outMap
			name = name + ""
			// log.Println(name, m.Name, m.RawName, m.Sig, hex.EncodeToString(m.ID))
		}
	}
}

func Inputs(data []byte) ([]interface{}, error) {
	if len(data) < 4 {
		return nil, errors.New("not found func sig")
	}
	method := hex.EncodeToString(data[:4])

	ms := FuncSigs[method]
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
	ms := FuncSigs[method]
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
		switch tv := v.(type) {
		case int:
			data[i] = big.NewInt(0).SetInt64(int64(tv))
		case int8:
			data[i] = big.NewInt(0).SetInt64(int64(tv))
		case int16:
			data[i] = big.NewInt(0).SetInt64(int64(tv))
		case int32:
			data[i] = big.NewInt(0).SetInt64(int64(tv))
		case int64:
			data[i] = big.NewInt(0).SetInt64(int64(tv))
		case uint:
			data[i] = big.NewInt(0).SetUint64(uint64(tv))
		case uint8:
			data[i] = big.NewInt(0).SetUint64(uint64(tv))
		case uint16:
			data[i] = big.NewInt(0).SetUint64(uint64(tv))
		case uint32:
			data[i] = big.NewInt(0).SetUint64(uint64(tv))
		case uint64:
			data[i] = big.NewInt(0).SetUint64(uint64(tv))
		}
	}

	var err error
	var bs []byte
	for _, m := range ms {
		bs, err = m.Outputs.Pack(data...)
		if err == nil {
			break
		}
	}
	return bs, err
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
