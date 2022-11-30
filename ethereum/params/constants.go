package params

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

var MaxUint256 = new(big.Int).Sub(new(big.Int).Lsh(common.Big1, 256), common.Big1)
