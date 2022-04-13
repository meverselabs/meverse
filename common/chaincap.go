package common

import (
	"math/big"
	"sync"
)

var chainCapMap map[uint64]*big.Int = map[uint64]*big.Int{}
var chainCapMapLock sync.Mutex

func GetChainCap(ChainID *big.Int) *big.Int {
	chainCapMapLock.Lock()
	ChainCap, has := chainCapMap[ChainID.Uint64()]
	chainCapMapLock.Unlock()
	if !has {
		ChainCap = big.NewInt(0).Mul(ChainID, big.NewInt(2))
		ChainCap.Add(ChainCap, big.NewInt(35))
		chainCapMapLock.Lock()
		chainCapMap[ChainID.Uint64()] = ChainCap
		chainCapMapLock.Unlock()
	}
	return ChainCap
}
