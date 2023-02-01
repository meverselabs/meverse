package erc20wrapper

import (
	"encoding/hex"
	"fmt"
	"math/big"

	ecommon "github.com/ethereum/go-ethereum/common"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/extern/txparser"
)

var (
	tagErc20Token = byte(0x01)
)

func addressToString32(addr common.Address) string {
	return "000000000000000000000000" + addr.String()[2:]
}

func funcSig(method string) string {
	m := txparser.Abi(method)
	return hex.EncodeToString(m.ID)
}

func toAmount(buf []byte) *amount.Amount {
	return &amount.Amount{Int: new(big.Int).SetBytes(buf)}
}

func unpackString(output []byte) (string, error) {
	begin, length, err := lengthPrefixPointsTo(0, output)
	if err != nil {
		return "", err
	}
	return string(output[begin : begin+length]), nil
}

// lengthPrefixPointsTo interprets a 32 byte slice as an offset and then determines which indices to look to decode the type.
func lengthPrefixPointsTo(index int, output []byte) (start int, length int, err error) {
	bigOffsetEnd := big.NewInt(0).SetBytes(output[index : index+32])
	bigOffsetEnd.Add(bigOffsetEnd, ecommon.Big32)
	outputLength := big.NewInt(int64(len(output)))

	if bigOffsetEnd.Cmp(outputLength) > 0 {
		return 0, 0, fmt.Errorf("abi: cannot marshal in to go slice: offset %v would go over slice boundary (len=%v)", bigOffsetEnd, outputLength)
	}

	if bigOffsetEnd.BitLen() > 63 {
		return 0, 0, fmt.Errorf("abi offset larger than int64: %v", bigOffsetEnd)
	}

	offsetEnd := int(bigOffsetEnd.Uint64())
	lengthBig := big.NewInt(0).SetBytes(output[offsetEnd-32 : offsetEnd])

	totalSize := big.NewInt(0)
	totalSize.Add(totalSize, bigOffsetEnd)
	totalSize.Add(totalSize, lengthBig)
	if totalSize.BitLen() > 63 {
		return 0, 0, fmt.Errorf("abi: length larger than int64: %v", totalSize)
	}

	if totalSize.Cmp(outputLength) > 0 {
		return 0, 0, fmt.Errorf("abi: cannot marshal in to go type: length insufficient %v require %v", outputLength, totalSize)
	}
	start = int(bigOffsetEnd.Uint64())
	length = int(lengthBig.Uint64())
	return
}
