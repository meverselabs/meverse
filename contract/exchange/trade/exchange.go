package trade

import (
	"math/big"
	"sync"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
	"github.com/pkg/errors"

	. "github.com/meverselabs/meverse/contract/exchange/util"
)

// Contract 가 아님
type Exchange struct {
	sync.Mutex
	addr   common.Address
	master common.Address
}

//////////////////////////////////////////////////
// Exchange : contract function
//////////////////////////////////////////////////
func (self *Exchange) Address() common.Address {

	return self.addr
}
func (self *Exchange) Master() common.Address {
	return self.master
}
func (self *Exchange) Init(addr common.Address, master common.Address) {
	self.addr = addr
	self.master = master
}
func (self *Exchange) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

//////////////////////////////////////////////////
// Exchange Contract : utility function
//////////////////////////////////////////////////
func (self *Exchange) makeSlice(cc types.ContractLoader) (uint8, []*big.Int) {
	N := self.nTokens(cc)
	result := MakeSlice(N)
	return N, result
}
func (self *Exchange) cloneSlice(cc types.ContractLoader, input []*big.Int) (uint8, []*big.Int) {
	N, result := self.makeSlice(cc)
	for i := uint8(0); i < N; i++ {
		result[i] = big.NewInt(0).Set(input[i])
	}
	return N, result
}

//////////////////////////////////////////////////
// Exchange Contract : modifier
//////////////////////////////////////////////////
func (self *Exchange) onlyOwner(cc *types.ContractContext) error {
	owner := self.owner(cc)
	if cc.From() != owner {
		return errors.New("Exchange: FORBIDDEN")
	}
	return nil
}

//////////////////////////////////////////////////
// Exchange : private reader functions
//////////////////////////////////////////////////
func (self *Exchange) exType(cc types.ContractLoader) uint8 {
	bs := cc.ContractData([]byte{tagExType})[0]
	return uint8(bs)
}
func (self *Exchange) factory(cc types.ContractLoader) common.Address {
	bs := cc.ContractData([]byte{tagFactory})
	return common.BytesToAddress(bs)
}
func (self *Exchange) owner(cc types.ContractLoader) common.Address {
	bs := cc.ContractData([]byte{tagOwner})
	return common.BytesToAddress(bs)
}
func (self *Exchange) winner(cc types.ContractLoader) common.Address {
	bs := cc.ContractData([]byte{tagExWinner})
	return common.BytesToAddress(bs)
}
func (self *Exchange) futureOwner(cc types.ContractLoader) common.Address {
	bs := cc.ContractData([]byte{tagExFutureOwner})
	return common.BytesToAddress(bs)
}
func (self *Exchange) futureWinner(cc types.ContractLoader) common.Address {
	bs := cc.ContractData([]byte{tagExFutureWinner})
	return common.BytesToAddress(bs)
}
func (self *Exchange) transferOwnerWinnerDeadline(cc types.ContractLoader) uint64 {
	bs := cc.ContractData([]byte{tagExTransferOwnerWinnerDeadline})
	if bs == nil || len(bs) == 0 {
		return uint64(0)
	}
	return bin.Uint64(bs)
}
func (self *Exchange) fee(cc *types.ContractContext) uint64 {
	bs := cc.ContractData([]byte{tagExFee})
	if bs == nil {
		return uint64(0)
	}
	return bin.Uint64(bs)
}
func (self *Exchange) futureFee(cc types.ContractLoader) uint64 {
	bs := cc.ContractData([]byte{tagExFutureFee})
	if bs == nil {
		return uint64(0)
	}
	return bin.Uint64(bs)
}
func (self *Exchange) adminFee(cc types.ContractLoader) uint64 {
	bs := cc.ContractData([]byte{tagExAdminFee})
	if bs == nil {
		return uint64(0)
	}
	return bin.Uint64(bs)
}
func (self *Exchange) futureAdminFee(cc types.ContractLoader) uint64 {
	bs := cc.ContractData([]byte{tagExFutureAdminFee})
	if bs == nil {
		return uint64(0)
	}
	return bin.Uint64(bs)
}
func (self *Exchange) winnerFee(cc types.ContractLoader) uint64 {
	bs := cc.ContractData([]byte{tagExWinnerFee})
	if bs == nil {
		return uint64(0)
	}
	return bin.Uint64(bs)
}
func (self *Exchange) futureWinnerFee(cc types.ContractLoader) uint64 {
	bs := cc.ContractData([]byte{tagExFutureWinnerFee})
	if bs == nil {
		return uint64(0)
	}
	return bin.Uint64(bs)
}
func (self *Exchange) adminActionsDeadline(cc types.ContractLoader) uint64 {
	bs := cc.ContractData([]byte{tagExAdminActionsDeadline})
	if bs == nil || len(bs) == 0 {
		return uint64(0)
	}
	return bin.Uint64(bs)
}
func (self *Exchange) whiteList(cc types.ContractLoader) common.Address {
	bs := cc.ContractData([]byte{tagExWhiteList})
	return common.BytesToAddress(bs)
}
func (self *Exchange) futureWhiteList(cc types.ContractLoader) common.Address {
	bs := cc.ContractData([]byte{tagExFutureWhiteList})
	return common.BytesToAddress(bs)
}
func (self *Exchange) groupId(cc types.ContractLoader) hash.Hash256 {
	bs := cc.ContractData([]byte{tagExGroupId})
	var h hash.Hash256
	copy(h[:], bs)
	return h
}
func (self *Exchange) futureGroupId(cc types.ContractLoader) hash.Hash256 {
	bs := cc.ContractData([]byte{tagExFutureGroupId})
	var h hash.Hash256
	copy(h[:], bs)
	return h
}
func (self *Exchange) whiteListDeadline(cc types.ContractLoader) uint64 {
	bs := cc.ContractData([]byte{tagExWhiteListDeadline})
	if bs == nil || len(bs) == 0 {
		return uint64(0)
	}
	return bin.Uint64(bs)
}

func (self *Exchange) feeWhiteList(cc *types.ContractContext, from common.Address) ([]byte, error) {
	/*
		if cc.IsContract(from) {
			return nil, errors.New("Exchange: CONTRACT")
		}
	*/
	whiteList := self.whiteList(cc)
	groupId := self.groupId(cc)
	is, err := cc.Exec(cc, whiteList, "GroupData", []interface{}{groupId, from})
	if err != nil {
		return nil, err
	}
	return is[0].([]byte), nil
}

// 1. from == ZeroAddress : fee()
// 2. from != ZeroAddress : 조회
// 2.1  from == whitelist  :  feeWhiteList(from)
// 2.2  from != whitelist  :  fee()
func (self *Exchange) feeAddress(cc *types.ContractContext, from common.Address) (uint64, error) {
	if from == ZeroAddress {
		return self.fee(cc), nil
	}
	bs, err := self.feeWhiteList(cc, from)
	if err != nil {
		return 0, err
	}
	if len(bs) != 0 {
		return bin.Uint64(bs), nil
	}
	return self.fee(cc), nil
}

// 1. cc.From() != contract, from == ZeroAddress  : feeAddress(cc.From())
// 2. cc.From() != contract, from != ZeroAddress  : feeAddress(from)
// 3. cc.From() == contract, from == ZeroAddress  : fee()
// 4. cc.From() == contract, from != ZeroAddress  : feeAddress(from)
func (self *Exchange) _feeAddress(cc *types.ContractContext, _from common.Address) (uint64, error) {

	if _from != ZeroAddress {
		return self.feeAddress(cc, _from)
	}

	if !cc.IsContract(cc.From()) {
		return self.feeAddress(cc, cc.From())
	} else {
		return self.fee(cc), nil
	}
}

func (self *Exchange) nTokens(cc types.ContractLoader) uint8 {
	bs := cc.ContractData([]byte{tagExNTokens})
	if bs == nil {
		return uint8(0)
	}
	return uint8(bs[0])
}
func (self *Exchange) tokens(cc types.ContractLoader) []common.Address {
	N := self.nTokens(cc)
	result := make([]common.Address, N, N)
	bs := cc.ContractData([]byte{tagExTokens})
	if len(bs) == 0 {
		for i := uint8(0); i < N; i++ {
			result[i] = ZeroAddress
		}
		return result
	}
	for i := uint8(0); i < N; i++ {
		result[i].SetBytes(bs[i*common.AddressLength : (i+1)*common.AddressLength])
	}

	return result
}
func (self *Exchange) payToken(cc types.ContractLoader) common.Address {
	bs := cc.ContractData([]byte{tagExPayToken})
	if bs == nil || len(bs) == 0 {
		return ZeroAddress
	}
	return common.BytesToAddress(bs)
}
func (self *Exchange) payTokenIndex(cc types.ContractLoader) (uint8, error) {
	N := self.nTokens(cc)
	tokens := self.tokens(cc)
	pt := self.payToken(cc)
	for i := uint8(0); i < N; i++ {
		if tokens[i] == pt {
			return i, nil
		}
	}
	return 255, errors.New("Exchange: PAY_INDEX")
}
func (self *Exchange) isKilled(cc types.ContractLoader) bool {
	bs := cc.ContractData([]byte{tagExIsKilled})
	if bs == nil || len(bs) == 0 || uint8(bs[0]) == 0 {
		return false
	}
	return true
}
func (self *Exchange) blockTimestampLast(cc types.ContractLoader) uint64 {
	bs := cc.ContractData([]byte{tagBlockTimestampLast})
	if bs == nil || len(bs) == 0 {
		return uint64(0)
	}
	return bin.Uint64(bs)
}

//////////////////////////////////////////////////
// Exchange : private writer Functions
//////////////////////////////////////////////////
func (self *Exchange) _setExtype(cc *types.ContractContext, _exType uint8) {
	cc.SetContractData([]byte{tagExType}, []byte{byte(_exType)})
}
func (self *Exchange) _setNTokens(cc *types.ContractContext, _nTokens uint8) error {
	if _nTokens < 2 {
		return errors.New("Exchange: TOKEN_NUMBER_LESSER_THAN_2")
	}
	cc.SetContractData([]byte{tagExNTokens}, []byte{_nTokens})
	return nil
}
func (self *Exchange) _setOwner(cc *types.ContractContext, _owner common.Address) error {
	if _owner == ZeroAddress {
		return errors.New("Exchange: ZERO_ADDRESS")
	}
	cc.SetContractData([]byte{tagOwner}, _owner.Bytes())
	return nil
}
func (self *Exchange) _setFutureOwner(cc *types.ContractContext, _owner common.Address) error {
	if _owner == ZeroAddress {
		return errors.New("Exchange: ZERO_ADDRESS")
	}
	cc.SetContractData([]byte{tagExFutureOwner}, _owner.Bytes())
	return nil
}
func (self *Exchange) _setWinner(cc *types.ContractContext, _owner common.Address) {
	cc.SetContractData([]byte{tagExWinner}, _owner.Bytes())
}
func (self *Exchange) _setFutureWinner(cc *types.ContractContext, _owner common.Address) {
	cc.SetContractData([]byte{tagExFutureWinner}, _owner.Bytes())
}
func (self *Exchange) _setAdminActionsDeadline(cc *types.ContractContext, _time uint64) {
	cc.SetContractData([]byte{tagExAdminActionsDeadline}, bin.Uint64Bytes(_time))
}
func (self *Exchange) _setPayToken(cc *types.ContractContext, _token common.Address) error {
	// 설정 없는 경우
	if _token == ZeroAddress {
		cc.SetContractData([]byte{tagExPayToken}, _token.Bytes())
		return nil
	}
	// 설정 있는 경우
	N := self.nTokens(cc)
	tokens := self.tokens(cc)
	for i := uint8(0); i < N; i++ {
		if tokens[i] == _token {
			cc.SetContractData([]byte{tagExPayToken}, _token.Bytes())
			return nil
		}
	}
	return errors.New("Exchange: NOT_EXIST_PAYTOKEN")
}
func (self *Exchange) setPayToken(cc *types.ContractContext, _token common.Address) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}
	return self._setPayToken(cc, _token)
}

func (self *Exchange) _setFee(cc *types.ContractContext, _fee uint64) error {
	if _fee > MAX_FEE {
		return errors.New("Exchange: FEE_EXCEED_MAXFEE")
	}
	cc.SetContractData([]byte{tagExFee}, bin.Uint64Bytes(_fee))
	return nil
}
func (self *Exchange) _setFutureFee(cc *types.ContractContext, _fee uint64) error {
	if _fee > MAX_FEE {
		return errors.New("Exchange: FUTURE_FEE_EXCEED_MAXFEE")
	}
	cc.SetContractData([]byte{tagExFutureFee}, bin.Uint64Bytes(_fee))
	return nil
}
func (self *Exchange) _setAdminFee(cc *types.ContractContext, _admin_fee uint64) error {
	if _admin_fee > MAX_ADMIN_FEE {
		return errors.New("Exchange: ADMIN_FEE_EXCEED_MAXADMINFEE")
	}
	cc.SetContractData([]byte{tagExAdminFee}, bin.Uint64Bytes(_admin_fee))
	return nil
}
func (self *Exchange) _setFutureAdminFee(cc *types.ContractContext, _admin_fee uint64) error {

	if _admin_fee > MAX_ADMIN_FEE {
		return errors.New("Exchange: FUTURE_ADMIN_FEE_EXCEED_MAXADMINFEE")
	}

	cc.SetContractData([]byte{tagExFutureAdminFee}, bin.Uint64Bytes(_admin_fee))
	return nil
}
func (self *Exchange) _setWinnerFee(cc *types.ContractContext, _winner_fee uint64) error {
	if _winner_fee > MAX_WINNER_FEE {
		return errors.New("Exchange: WINNER_FEE_EXCEED_MAXADMINFEE")
	}
	cc.SetContractData([]byte{tagExWinnerFee}, bin.Uint64Bytes(_winner_fee))
	return nil
}
func (self *Exchange) _setFutureWinnerFee(cc *types.ContractContext, _winner_fee uint64) error {
	if _winner_fee > MAX_WINNER_FEE {
		return errors.New("Exchange: FUTURE_WINNER_FEE_EXCEED_MAXADMINFEE")
	}
	cc.SetContractData([]byte{tagExFutureWinnerFee}, bin.Uint64Bytes(_winner_fee))
	return nil
}
func (self *Exchange) _setWhiteListDeadline(cc *types.ContractContext, _time uint64) {
	cc.SetContractData([]byte{tagExWhiteListDeadline}, bin.Uint64Bytes(_time))
}
func (self *Exchange) _setWhiteList(cc *types.ContractContext, _whiteList common.Address) {
	cc.SetContractData([]byte{tagExWhiteList}, _whiteList.Bytes())
}
func (self *Exchange) _setFutureWhiteList(cc *types.ContractContext, _whiteList common.Address) {
	cc.SetContractData([]byte{tagExFutureWhiteList}, _whiteList.Bytes())
}
func (self *Exchange) _setGroupId(cc *types.ContractContext, _groupId hash.Hash256) {
	cc.SetContractData([]byte{tagExGroupId}, _groupId[:])
}
func (self *Exchange) _setFutureGroupId(cc *types.ContractContext, _groupId hash.Hash256) {
	cc.SetContractData([]byte{tagExFutureGroupId}, _groupId[:])
}
func (self *Exchange) commitNewFee(cc *types.ContractContext, _new_fee, _new_admin_fee, _new_winner_fee, delay uint64) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}
	if !(self.adminActionsDeadline(cc) == 0) { // active action
		return errors.New("Exchange: ADMIN_ACTIONS_DEADLINE")
	}

	deadline := cc.LastTimestamp()/uint64(time.Second) + delay
	self._setAdminActionsDeadline(cc, deadline)
	if err := self._setFutureFee(cc, _new_fee); err != nil {
		return err
	}
	if err := self._setFutureAdminFee(cc, _new_admin_fee); err != nil {
		return err
	}
	if err := self._setFutureWinnerFee(cc, _new_winner_fee); err != nil {
		return err
	}
	return nil
}

func (self *Exchange) applyNewFee(cc *types.ContractContext) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}
	if (cc.LastTimestamp() / uint64(time.Second)) < self.adminActionsDeadline(cc) {
		return errors.New("Exchange: ADMIN_ACTIONS_DEADLINE") // insufficient time
	}
	if self.adminActionsDeadline(cc) == 0 {
		return errors.New("Exchange: NO_ACTIVE_ACTION") // no active action
	}

	self._setAdminActionsDeadline(cc, 0)

	self._setFee(cc, self.futureFee(cc))
	self._setAdminFee(cc, self.futureAdminFee(cc))
	self._setWinnerFee(cc, self.futureWinnerFee(cc))
	return nil
}
func (self *Exchange) revertNewFee(cc *types.ContractContext) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}
	self._setAdminActionsDeadline(cc, 0)
	return nil
}

func (self *Exchange) commitNewWhiteList(cc *types.ContractContext, new_whiteList common.Address, new_groupId hash.Hash256, delay uint64) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}
	if !(self.whiteListDeadline(cc) == 0) { // active action
		return errors.New("Exchange: WHITELIST_DEADLINE")
	}

	deadline := cc.LastTimestamp()/uint64(time.Second) + delay
	self._setWhiteListDeadline(cc, deadline)
	self._setFutureWhiteList(cc, new_whiteList)
	self._setFutureGroupId(cc, new_groupId)
	return nil
}

func (self *Exchange) applyNewWhiteList(cc *types.ContractContext) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}
	if (cc.LastTimestamp() / uint64(time.Second)) < self.whiteListDeadline(cc) {
		return errors.New("Exchange: WHITELIST_DEADLINE") // insufficient time
	}
	if self.whiteListDeadline(cc) == 0 {
		return errors.New("Exchange: NO_ACTIVE_ACTION") // no active action
	}

	self._setWhiteListDeadline(cc, 0)
	self._setWhiteList(cc, self.futureWhiteList(cc))
	self._setGroupId(cc, self.futureGroupId(cc))
	return nil
}
func (self *Exchange) revertNewWhiteList(cc *types.ContractContext) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}
	self._setWhiteListDeadline(cc, 0)
	return nil
}

// onlyOwner,  divide adminFee to owner and winner according to winnerfee
// not token0, token1 but tokens[]
func (self *Exchange) _divideFee(cc *types.ContractContext, _adminFees []*big.Int) ([]*big.Int, []*big.Int, error) {
	N := self.nTokens(cc)
	tokens := self.tokens(cc)
	owner := self.owner(cc)
	winner := self.winner(cc)
	winnerFeeNominator := big.NewInt(int64(self.winnerFee(cc)))

	ownerFees := MakeSlice(N)
	winnerFees := MakeSlice(N)
	if winner != ZeroAddress && winnerFeeNominator.Cmp(Zero) != 0 {
		for i := uint8(0); i < N; i++ {
			fee := _adminFees[i]
			if fee.Cmp(Zero) == 0 {
				continue
			}
			winnerFees[i].Set(MulDivC(fee, winnerFeeNominator, FEE_DENOMINATOR))
			ownerFees[i].Set(Sub(fee, winnerFees[i]))
		}
	} else {
		ownerFees = CloneSlice(_adminFees)
	}

	for i := uint8(0); i < N; i++ {
		if ownerFees[i].Cmp(Zero) > 0 { // owner 지급
			if err := SafeTransfer(cc, tokens[i], owner, ownerFees[i]); err != nil {
				return nil, nil, err
			}
		}
		if winnerFees[i].Cmp(Zero) > 0 { // winner 지급
			if err := SafeTransfer(cc, tokens[i], winner, winnerFees[i]); err != nil {
				return nil, nil, err
			}
		}
	}
	return ownerFees, winnerFees, nil
}
func (self *Exchange) _setTransferOwnerWinnerDeadline(cc *types.ContractContext, _time uint64) {
	cc.SetContractData([]byte{tagExTransferOwnerWinnerDeadline}, bin.Uint64Bytes(_time))
}
func (self *Exchange) commitTransferOwnerWinner(cc *types.ContractContext, _new_owner, _new_winner common.Address, delay uint64) error {
	if err := self.onlyOwner(cc); err != nil { //  only owner
		return err
	}
	if self.transferOwnerWinnerDeadline(cc) != 0 {
		return errors.New("Exchange: ACTIVE_TRANSFER") //  active transfer
	}

	deadline := cc.LastTimestamp()/uint64(time.Second) + delay
	self._setTransferOwnerWinnerDeadline(cc, deadline)
	if err := self._setFutureOwner(cc, _new_owner); err != nil {
		return err
	}
	self._setFutureWinner(cc, _new_winner)

	return nil
}
func (self *Exchange) _applyTransferOwnerWinner(cc *types.ContractContext) (common.Address, common.Address, error) {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return ZeroAddress, ZeroAddress, err
	}
	if cc.LastTimestamp()/uint64(time.Second) < self.transferOwnerWinnerDeadline(cc) {
		return ZeroAddress, ZeroAddress, errors.New("Exchange: INSUFFICIENT_TIME")
	}
	if self.transferOwnerWinnerDeadline(cc) == 0 {
		return ZeroAddress, ZeroAddress, errors.New("Exchange: NO_ACTIVE_TRANSFER")
	}

	self._setTransferOwnerWinnerDeadline(cc, 0)
	futureOwner := self.futureOwner(cc)
	if err := self._setOwner(cc, futureOwner); err != nil {
		return ZeroAddress, ZeroAddress, err
	}
	self._setWinner(cc, self.futureWinner(cc))

	return cc.From(), futureOwner, nil // old owner, new owner
}
func (self *Exchange) revertTransferOwnerWinner(cc *types.ContractContext) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}
	self._setTransferOwnerWinnerDeadline(cc, 0)
	return nil
}
func (self *Exchange) _setIsKilled(cc *types.ContractContext, _killed bool) error {
	if err := self.onlyOwner(cc); err != nil { // only Owner
		return err
	}
	b := uint8(0)
	if _killed {
		b = uint8(1)
	}
	cc.SetContractData([]byte{tagExIsKilled}, []byte{byte(b)})
	return nil
}
func (self *Exchange) killMe(cc *types.ContractContext) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}
	self._setIsKilled(cc, true)
	return nil
}
func (self *Exchange) unkillMe(cc *types.ContractContext) error {
	if err := self.onlyOwner(cc); err != nil { // only owner
		return err
	}
	self._setIsKilled(cc, false)
	return nil
}
func (self *Exchange) _setBlockTimestampLast(cc *types.ContractContext, blockTimestampLast uint64) {
	cc.SetContractData([]byte{tagBlockTimestampLast}, bin.Uint64Bytes(blockTimestampLast))
}

// tokenTranfer : emergency use
func (self *Exchange) tokenTransfer(cc *types.ContractContext, token, to common.Address, amt *amount.Amount) error {
	if err := self.onlyOwner(cc); err != nil { //  only owner
		return err
	}

	N := self.nTokens(cc)
	tokens := self.tokens(cc)

	for i := uint8(0); i < N; i++ {
		if tokens[i] == token {
			_, err := cc.Exec(cc, token, "Transfer", []interface{}{to, amt})
			return err
		}
	}

	return errors.New("Exchange: NOT_EXIST_TOKEN")
}
