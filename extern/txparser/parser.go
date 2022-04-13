package txparser

import (
	"bytes"
	"encoding/hex"

	"github.com/ethereum/go-ethereum/accounts/abi"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/fletaio/fleta_v2/common/amount"
	"github.com/pkg/errors"
	"golang.org/x/crypto/sha3"
)

var (
	FuncSigs          map[string]abi.Method
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
	FuncSigs = map[string]abi.Method{}
	ercs := [][]byte{
		IERC20,
		IERC721,
		IERC1155,
		Formulator,
	}
	for _, v := range ercs {
		reader := bytes.NewReader(v)
		a, err := abi.JSON(reader)
		if err != nil {
			panic(err)
		}
		for _, m := range a.Methods {
			// log.Println(name, m.Name, m.RawName, m.Sig)
			FuncSigs[m.Sig] = m
			transferFnSignature := []byte(m.Sig)
			hash := sha3.NewLegacyKeccak256()
			hash.Write(transferFnSignature)

			FuncSigs[hex.EncodeToString(hash.Sum(nil)[:4])] = m
		}
	}
}

func Inputs(data []byte) ([]interface{}, error) {
	if len(data) < 4 {
		return nil, errors.New("not found func sig")
	}
	method := hex.EncodeToString(data[:4])

	m := FuncSigs[method]
	obj, err := m.Inputs.Unpack(data[4:])
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func Outputs(method string, data []interface{}) ([]byte, error) {
	m := FuncSigs[method]
	for i, v := range data {
		if a, ok := v.(*amount.Amount); ok {
			data[i] = a.Int
		}
	}
	bs, err := m.Outputs.Pack(data...)
	if err != nil {
		return nil, err
	}
	return bs, nil
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
	t.UnmarshalBinary(rlpBytes)
	etx, err := &t, t.UnmarshalBinary(rlpBytes)
	if err != nil {
		return nil, nil, err
	}
	v, r, s := etx.RawSignatureValues()

	sig := []byte{}
	sig = append(sig, r.Bytes()...)
	sig = append(sig, s.Bytes()...)
	sig = append(sig, v.Bytes()...)

	return etx, sig, nil
}
