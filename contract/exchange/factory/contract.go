package factory

import (
	"bytes"
	"math/big"

	"github.com/pkg/errors"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"

	"github.com/meverselabs/meverse/contract/exchange/trade"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/meverselabs/meverse/contract/exchange/util"
)

type FactoryContract struct {
	addr   common.Address
	master common.Address
}

func (cont *FactoryContract) Address() common.Address {
	return cont.addr
}
func (cont *FactoryContract) Master() common.Address {
	return cont.master
}
func (cont *FactoryContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}
func (cont *FactoryContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &FactoryContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}

	cc.SetContractData([]byte{tagOwner}, data.Owner[:])

	return nil
}
func (cont *FactoryContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}
func (cont *FactoryContract) owner(cc types.ContractLoader) common.Address {
	bs := cc.ContractData([]byte{tagOwner})
	return common.BytesToAddress(bs)
}
func (cont *FactoryContract) getPair(cc types.ContractLoader, token0, token1 common.Address) common.Address {
	pair := cc.ContractData(makePairKey(token0, token1))
	if pair == nil {
		return ZeroAddress
	}
	return common.BytesToAddress(pair)
}
func (cont *FactoryContract) allPairs(cc types.ContractLoader) []common.Address {
	bs := cc.ContractData([]byte{tagAllPairs})
	if bs == nil {
		return nil
	}

	addr := ZeroAddress
	allPairs := []common.Address{}

	for i := int(0); i < len(bs); i += common.AddressLength {
		copy(addr[0:], bs[i:i+common.AddressLength])
		allPairs = append(allPairs, addr)
	}

	return allPairs
}
func (cont *FactoryContract) allPairsLength(cc types.ContractLoader) uint16 {
	bs := cc.ContractData([]byte{tagAllPairs})
	if bs == nil {
		return uint16(0)
	}
	return uint16(len(bs) / common.AddressLength)
}
func (cont *FactoryContract) _addresses(cc *types.ContractContext, tokenA, tokenB common.Address) (common.Address, common.Address, common.Address, error) {
	err := cont.onlyOwner(cc)
	if err != nil {
		return ZeroAddress, ZeroAddress, ZeroAddress, err
	}

	token0, token1, err := trade.SortTokens(tokenA, tokenB)
	if err != nil {
		return ZeroAddress, ZeroAddress, ZeroAddress, err
	}
	// 기존에 존재하면
	pairKey := makePairKey(token0, token1)
	if cc.ContractData(pairKey) != nil {
		return ZeroAddress, ZeroAddress, ZeroAddress, errors.New("Exchange: PAIR_EXISTS")
	}
	// Contract Deploy
	pair, err := trade.PairFor(cont.addr, token0, token1) //trade Address
	if err != nil {
		return ZeroAddress, ZeroAddress, ZeroAddress, err
	}

	return pair, token0, token1, nil
}

// 데이터 저장
func (cont *FactoryContract) _setData(cc *types.ContractContext, pair, token0, token1 common.Address) {

	cc.SetContractData(makePairKey(token0, token1), pair.Bytes())
	cc.SetContractData(makePairKey(token1, token0), pair.Bytes())

	bs := cc.ContractData([]byte{tagAllPairs})
	if bs == nil {
		bs = []byte{}
	}

	bs = append(bs, pair.Bytes()...)
	cc.SetContractData([]byte{tagAllPairs}, bs)
}
func (cont *FactoryContract) createPairUni(cc *types.ContractContext, tokenA, tokenB, _payToken common.Address, name, symbol string, _owner, _winner common.Address, _fee, _adminFee, _winnerFee uint64, _whiteList common.Address, _groupId hash.Hash256, classID uint64) (common.Address, error) {
	if err := cont.onlyOwner(cc); err != nil {
		return ZeroAddress, err
	}

	pair, token0, token1, err := cont._addresses(cc, tokenA, tokenB)
	if err != nil {
		return ZeroAddress, err
	}
	pairConstrunction := &trade.UniSwapConstruction{
		Name:      name,
		Symbol:    symbol,
		Factory:   cont.addr,
		Token0:    token0,
		Token1:    token1,
		PayToken:  _payToken,
		Owner:     _owner,
		Winner:    _winner,
		Fee:       _fee,
		AdminFee:  _adminFee,
		WinnerFee: _winnerFee,
		WhiteList: _whiteList,
		GroupId:   _groupId,
	}
	bs, _, err := bin.WriterToBytes(pairConstrunction)
	if err != nil {
		return ZeroAddress, err
	}
	if _, err = cc.DeployContractWithAddress(cont.addr, classID, pair, bs); err != nil {
		return ZeroAddress, err
	}

	cont._setData(cc, pair, token0, token1)

	return pair, nil
}
func (cont *FactoryContract) createPairStable(cc *types.ContractContext, tokenA, tokenB, _payToken common.Address, name, symbol string, _owner, _winner common.Address, _fee, _adminFee, _winnerFee uint64, _whiteList common.Address, _groupId hash.Hash256, _amp uint64, classID uint64) (common.Address, error) {
	if err := cont.onlyOwner(cc); err != nil {
		return ZeroAddress, err
	}

	pair, token0, token1, err := cont._addresses(cc, tokenA, tokenB)
	if err != nil {
		return ZeroAddress, err
	}

	// 지정하는 순서대로 -  not token0, token1
	tokens := []common.Address{tokenA, tokenB}

	pairConstrunction := &trade.StableSwapConstruction{
		Name:         name,
		Symbol:       symbol,
		Factory:      cont.addr,
		NTokens:      uint8(2),
		Tokens:       tokens,
		PayToken:     _payToken,
		Owner:        _owner,
		Winner:       _winner,
		Fee:          _fee,
		AdminFee:     _adminFee,
		WinnerFee:    _winnerFee,
		WhiteList:    _whiteList,
		GroupId:      _groupId,
		Amp:          big.NewInt(int64(_amp)),
		PrecisionMul: []uint64{1, 1},
		Rates:        []*big.Int{big.NewInt(trade.PRECISION), big.NewInt(trade.PRECISION)},
	}

	bs, _, err := bin.WriterToBytes(pairConstrunction)
	if err != nil {
		return ZeroAddress, nil
	}

	if _, err = cc.DeployContractWithAddress(cont.addr, classID, pair, bs); err != nil {
		return ZeroAddress, err
	}

	cont._setData(cc, pair, token0, token1)

	return pair, nil
}
func (cont *FactoryContract) onlyOwner(cc *types.ContractContext) error {
	owner := common.BytesToAddress(cc.ContractData([]byte{tagOwner}))
	if cc.From() != owner {
		return errors.New("Exchange: FORBIDDEN")
	}
	return nil
}
func (cont *FactoryContract) setOwner(cc *types.ContractContext, _owner common.Address) error {
	if err := cont.onlyOwner(cc); err != nil {
		return err
	}
	cc.SetContractData([]byte{tagOwner}, _owner.Bytes())
	return nil
}
