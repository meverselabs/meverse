package types

import (
	"math/big"

	etypes "github.com/ethereum/go-ethereum/core/types"
)

// MakeSigner returns a Signer based on the given chain config and block number.
// 현재는 LondonSigner만 생기나, 앞으로 다른 signer가 생길 수 있음
func MakeSigner(chainID *big.Int, blockNumber uint32) etypes.Signer {
	var signer etypes.Signer

	signer = etypes.NewLondonSigner(chainID)
	return signer
}
