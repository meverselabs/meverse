package bridge

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/core/types"
)

type BridgeContract struct {
	addr   common.Address
	master common.Address
}

func (cont *BridgeContract) Name() string {
	return "BridgeContract"
}

func (cont *BridgeContract) Address() common.Address {
	return cont.addr
}

func (cont *BridgeContract) Master() common.Address {
	return cont.master
}

func (cont *BridgeContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

func (cont *BridgeContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &BridgeContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}
	cc.SetContractData([]byte{tagBankAddress}, data.Bank[:])
	cc.SetContractData([]byte{tagFeeOwnerAddress}, data.FeeOwner[:])
	cc.SetContractData([]byte{tagMeverseTokenAddress}, data.MeverseToken[:])

	return nil
}

func (cont *BridgeContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

var feeFactorMax uint16 = 10000

//////////////////////////////////////////////////
// Private Functions
//////////////////////////////////////////////////

func (cont *BridgeContract) addSequenceFrom(cc *types.ContractContext, user common.Address, chain string) {
	bs := cc.ContractData(makeSequenceFrom(user, chain))
	bi := big.NewInt(0).SetBytes(bs)
	bi.Add(bi, big.NewInt(1))
	cc.SetContractData(makeSequenceFrom(user, chain), bi.Bytes())
}

func (cont *BridgeContract) addSequenceTo(cc *types.ContractContext, user common.Address, chain string) {
	bs := cc.ContractData(makeSequenceTo(user, chain))
	bi := big.NewInt(0).SetBytes(bs)
	bi.Add(bi, big.NewInt(1))
	cc.SetContractData(makeSequenceTo(user, chain), bi.Bytes())
}

func getAllowance(cc *types.ContractContext, token common.Address, user common.Address, spander common.Address) (*amount.Amount, error) {
	if bal, err := cc.Exec(cc, token, "Allowance", []interface{}{user, spander}); err != nil {
		return nil, err
	} else if len(bal) < 1 {
		return nil, errors.New("invalid result balanceOf")
	} else {
		balAmt, ok := bal[0].(*amount.Amount)
		if !ok {
			return nil, errors.New("balanceof not valid response")
		}
		return balAmt, nil
	}
}

func getBalanceOf(cc *types.ContractContext, token common.Address, user common.Address) (*amount.Amount, error) {
	if bal, err := cc.Exec(cc, token, "BalanceOf", []interface{}{user}); err != nil {
		return nil, err
	} else if len(bal) < 1 {
		return nil, errors.New("invalid result balanceOf")
	} else {
		balAmt, ok := bal[0].(*amount.Amount)
		if !ok {
			return nil, errors.New("balanceof not valid response")
		}
		return balAmt, nil
	}
}

func safeTransfer(cc *types.ContractContext, token common.Address, To common.Address, amt *amount.Amount) error {
	if _, err := cc.Exec(cc, token, "Transfer", []interface{}{To, amt}); err != nil {
		return err
	}
	return nil
}

func safeTransferFrom(cc *types.ContractContext, token common.Address, From common.Address, To common.Address, amt *amount.Amount) error {
	if _, err := cc.Exec(cc, token, "TransferFrom", []interface{}{From, To, amt}); err != nil {
		return err
	}
	return nil
}

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////
func (cont *BridgeContract) sendToGateway(cc *types.ContractContext, token common.Address, amt *amount.Amount, path []common.Address, toChain string, summary []byte) error {
	// uint256 transferFee = msg.value;
	// require(
	// 	transferFee == transferFeeInfoToChain[toChain],
	// 	"sendToGateway: fee is not valid"
	// );
	if !amt.IsPlus() {
		return errors.New("sendToGateway: plus amount")
	}
	mt := *cc.MainToken()
	transferFee := cont.transferFeeInfoToChain(cc, toChain)
	transferTokenFeeFactor := cont.transferTokenFeeInfoToChain(cc, toChain)
	transferTokenFee := amt.MulC(int64(transferTokenFeeFactor)).DivC(int64(feeFactorMax))
	if transferTokenFee.IsMinus() {
		return errors.New("sendToGateway: invalid token fee")
	}

	// if bal, err := getBalanceOf(cc, *mt, cc.From()); err != nil {
	// 	return err
	// } else if bal.Cmp(transferFee.Int) < 0 {
	// 	return errors.New("sendToGateway: fee is not valid")
	// }
	// require(
	// 	IERC20(token).allowance(_msgSender(), address(this)) >= amt,
	// 	"sendToGateway: insufficient allowance"
	// );
	// if allow, err := getAllowance(cc, token, cc.From(), cont.addr); err != nil {
	// 	return err
	// } else if allow.Cmp(amt.Int) < 0 {
	// 	return errors.New("sendToGateway: insufficient allowance")
	// }
	// SafeERC20.safeTransferFrom(
	// 	IERC20(token),
	// 	_msgSender(),
	// 	address(this),
	// 	amt
	// );
	if err := safeTransferFrom(cc, token, cc.From(), cont.addr, amt); err != nil {
		return err
	}

	// if (transferFee > 0) {
	// 	payable(feeOwner()).transfer(transferFee);
	// }
	if transferFee.IsPlus() {
		if err := safeTransferFrom(cc, mt, cc.From(), cont.feeOwner(cc), transferFee); err != nil {
			return err
		}
	}
	tokenFeeOwner := cont.tokenFeeOwner(cc)
	if transferTokenFee.IsPlus() && tokenFeeOwner != common.ZeroAddr {
		if err := safeTransferFrom(cc, token, cc.From(), tokenFeeOwner, transferTokenFee); err != nil {
			return err
		}
	}
	// getSequenceFrom[_msgSender()][toChain]++;
	cont.addSequenceFrom(cc, cc.From(), toChain)
	return nil
}

//////////////////////////////////////////////////
// Public Read Functions
//////////////////////////////////////////////////

func (cont *BridgeContract) meverseToken(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagMeverseTokenAddress})
	var addr common.Address
	copy(addr[:], bs)
	return addr
}

func (cont *BridgeContract) bank(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagBankAddress})
	var addr common.Address
	copy(addr[:], bs)
	return addr
}

func (cont *BridgeContract) feeOwner(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagFeeOwnerAddress})
	var addr common.Address
	copy(addr[:], bs)
	return addr
}

func (cont *BridgeContract) tokenFeeOwner(cc *types.ContractContext) common.Address {
	bs := cc.ContractData([]byte{tagTokenFeeOwnerAddress})
	var addr common.Address
	copy(addr[:], bs)
	return addr
}

func (cont *BridgeContract) transferFeeInfoToChain(cc *types.ContractContext, chain string) *amount.Amount {
	bs := cc.ContractData(makeTransferFeeInfoToChain(chain))
	return amount.NewAmountFromBytes(bs)
}

func (cont *BridgeContract) transferTokenFeeInfoToChain(cc *types.ContractContext, chain string) uint16 {
	bs := cc.ContractData(makeTransferTokenFeeInfoToChain(chain))
	if len(bs) == 0 || len(bs) != 2 {
		return 0
	}
	return bin.Uint16(bs)
}

func (cont *BridgeContract) tokenFeeInfoFromChain(cc *types.ContractContext, chain string) uint16 {
	bs := cc.ContractData(makeTokenFeeInfoFromChain(chain))
	if len(bs) == 0 || len(bs) != 2 {
		return 0
	}
	return bin.Uint16(bs)
}

func (cont *BridgeContract) getSequenceFrom(cc *types.ContractContext, user common.Address, chain string) *big.Int {
	bs := cc.ContractData(makeSequenceFrom(user, chain))
	bi := big.NewInt(0).SetBytes(bs)
	return bi
}

func (cont *BridgeContract) getSequenceTo(cc *types.ContractContext, user common.Address, chain string) *big.Int {
	bs := cc.ContractData(makeSequenceTo(user, chain))
	bi := big.NewInt(0).SetBytes(bs)
	return bi
}

func (cont *BridgeContract) allowanceTokenFromGateway(cc *types.ContractContext, token common.Address, from common.Address) (*amount.Amount, error) {
	// return IERC20(token).allowance(from, address(this))
	return getAllowance(cc, token, from, cont.addr)
}

func (cont *BridgeContract) balanceOfToGateway(cc *types.ContractContext, token common.Address, from common.Address) (*amount.Amount, error) {
	return getBalanceOf(cc, token, from)
}

func (cont *BridgeContract) stringToBytes32(source string) []byte {
	return []byte(source)
}

//////////////////////////////////////////////////
// Public Writer only banker Functions
//////////////////////////////////////////////////
func (cont *BridgeContract) sendFromGateway(cc *types.ContractContext, token common.Address, to common.Address, amt *amount.Amount, path []common.Address, fromChain string, summary []byte) error {
	if cc.TargetHeight() > 11492000 {
		banker := cont.bank(cc)
		if cc.From() != banker {
			return errors.New("not banker")
		}
	}
	// uint256 amountChangedDecimal = getTokenAmount(fromChain, token, amount);
	// require(
	// 	IERC20(token).balanceOf(address(this)) >= amountChangedDecimal,
	// 	"sendFromGateway: insufficient contract balance"
	// );
	if bal, err := getBalanceOf(cc, token, cont.addr); err != nil {
		return err
		// require(
		// 	IERC20(token).balanceOf(address(this)) >= amountChangedDecimal,
		// 	"sendFromGateway: insufficient contract balance"
		// );
	} else if bal.Cmp(amt.Int) < 0 {
		return errors.New("sendFromGateway: insufficient contract balance")
	} else {
		// uint256 tokenFee = 0;
		// if (equals(fromChain, "MEVERSE") && meverseToken() != token) {
		// 	tokenFee = amountChangedDecimal * (tokenFeeInfoFromChain[fromChain] / feeFactorMax);
		// 	SafeERC20.safeTransfer(IERC20(token), to, tokenFee);
		// }
		// uint256 caldAmount = SafeMath.sub(amountChangedDecimal, tokenFee);
		// tokenFee := amount.NewAmount(0, 0)
		// if fromChain == "MEVERSE" && cont.meverseToken(cc) != token {
		tokenFeeFactor := cont.tokenFeeInfoFromChain(cc, fromChain)
		tokenFee := amt.MulC(int64(tokenFeeFactor)).DivC(int64(feeFactorMax))
		if tokenFee.IsPlus() {
			safeTransfer(cc, token, cont.feeOwner(cc), tokenFee)
		} else if tokenFee.IsMinus() {
			return errors.New("sendFromGateway: invalid token fee")
		}
		// }
		caldAmount := amt.Sub(tokenFee)

		// SafeERC20.safeTransfer(IERC20(token), to, caldAmount);
		err = safeTransfer(cc, token, to, caldAmount)
		if err != nil {
			return err
		}
		// getSequenceTo[to][fromChain]++;
		cont.addSequenceTo(cc, to, fromChain)
	}
	err := cont.sendMainToken(cc, fromChain, to, amt)
	if err != nil {
		return err
	}
	return nil
}

//////////////////////////////////////////////////
// Public Writer only owner Functions
//////////////////////////////////////////////////

func (cont *BridgeContract) checkOwner(cc *types.ContractContext) bool {
	if cc.TargetHeight() > 10780198 {
		return cc.From() == cont.Master()
	}
	return true
}

func (cont *BridgeContract) setTransferFeeInfo(cc *types.ContractContext, chain string, transferFee *amount.Amount) error {
	if !cont.checkOwner(cc) {
		return errors.New("not owner")
	}
	// transferFeeInfoToChain[chain] = transferFee;
	cc.SetContractData(makeTransferFeeInfoToChain(chain), transferFee.Bytes())
	return nil
}

func (cont *BridgeContract) setTransferTokenFeeInfo(cc *types.ContractContext, chain string, tokenFee uint16) error {
	if !cont.checkOwner(cc) {
		return errors.New("not owner")
	}
	cc.SetContractData(makeTransferTokenFeeInfoToChain(chain), bin.Uint16Bytes(tokenFee))
	return nil
}

func (cont *BridgeContract) setTokenFeeInfo(cc *types.ContractContext, chain string, tokenFee uint16) error {
	if !cont.checkOwner(cc) {
		return errors.New("not owner")
	}
	cc.SetContractData(makeTokenFeeInfoFromChain(chain), bin.Uint16Bytes(tokenFee))
	return nil
}

func (cont *BridgeContract) transferBankOwnership(cc *types.ContractContext, newBank common.Address) error {
	if !cont.checkOwner(cc) {
		return errors.New("not owner")
	}
	// require(
	// 	newBank != address(0),
	// 	"Bankable: new banker is the zero address"
	// );
	if newBank == common.ZeroAddr {
		return errors.New("bankable: new banker is the zero address")
	}
	cc.SetContractData([]byte{tagBankAddress}, newBank[:])
	return nil
}

func (cont *BridgeContract) changeMeverseAddress(cc *types.ContractContext, newTokenAddress common.Address) error {
	if !cont.checkOwner(cc) {
		return errors.New("not owner")
	}
	// require(
	// 	newTokenAddress != address(0),
	// 	"FeeOwnable: new FeeOwner is the zero address"
	// );
	if newTokenAddress == common.ZeroAddr {
		return errors.New("MeverseAddress: new MeverseAddress is the zero address")
	}
	cc.SetContractData([]byte{tagMeverseTokenAddress}, newTokenAddress[:])
	return nil
}

func (cont *BridgeContract) transferFeeOwnership(cc *types.ContractContext, newFeeOwner common.Address) error {
	if !cont.checkOwner(cc) {
		return errors.New("not owner")
	}
	// require(
	// 	newFeeOwner != address(0),
	// 	"FeeOwnable: new FeeOwner is the zero address"
	// );
	if newFeeOwner == common.ZeroAddr {
		return errors.New("FeeOwnable: new FeeOwner is the zero address")
	}
	cc.SetContractData([]byte{tagFeeOwnerAddress}, newFeeOwner[:])
	return nil
}

func (cont *BridgeContract) transferTokenFeeOwnership(cc *types.ContractContext, newFeeOwner common.Address) error {
	if !cont.checkOwner(cc) {
		return errors.New("not owner")
	}
	// require(
	// 	newFeeOwner != address(0),
	// 	"FeeOwnable: new FeeOwner is the zero address"
	// );
	if newFeeOwner == common.ZeroAddr {
		return errors.New("FeeOwnable: new FeeOwner is the zero address")
	}
	cc.SetContractData([]byte{tagTokenFeeOwnerAddress}, newFeeOwner[:])
	return nil
}

func (cont *BridgeContract) reclaimToken(cc *types.ContractContext, token common.Address, amt *amount.Amount) error {
	if !cont.checkOwner(cc) {
		return errors.New("not owner")
	}
	// uint256 balance = IERC20(token).balanceOf(address(this));
	if bal, err := getBalanceOf(cc, token, cont.addr); err != nil {
		return err
		// require(
		// 	balance >= amount,
		// 	"reclaimToken: insufficient contract balance"
		// );
	} else if bal.Cmp(amt.Int) < 0 {
		return errors.New("reclaimToken: insufficient contract balance")
	}

	// SafeERC20.safeTransfer(IERC20(token), _msgSender(), amount);
	return safeTransfer(cc, token, cc.From(), amt)
}

func (cont *BridgeContract) setSendMaintoken(cc *types.ContractContext, store common.Address, fromChains []string, overthens, amts []*amount.Amount) error {
	if !cont.checkOwner(cc) {
		return errors.New("not owner")
	}
	flag := cc.ContractData([]byte{tagSendMaintokenFlag})
	if len(flag) == 0 {
		cc.SetContractData([]byte{tagSendMaintokenFlag}, []byte{1})
	} else {
		removeSendMainTokenChains(cc)
	}

	if len(fromChains) != len(overthens) || len(fromChains) != len(amts) {
		return fmt.Errorf("not match params %v %v %v", len(fromChains), len(overthens), len(amts))
	}

	cc.SetContractData([]byte{tagMaintokenStore}, store[:])

	for i, fromChain := range fromChains {
		si := &SendMaintokenInfo{
			overthen: overthens[i],
			amt:      amts[i],
		}
		var buf bytes.Buffer
		_, err := si.WriteTo(&buf)
		if err != nil {
			return err
		}
		cc.SetContractData(makeSendMaintokenInfoKey(fromChain), buf.Bytes())
	}
	cc.SetContractData([]byte{tagSendChainList}, []byte(strings.Join(fromChains, ";")))

	// store common.Address, fromChain string, overthen, amt *amount.Amount
	return nil
}

func (cont *BridgeContract) sendMainToken(cc *types.ContractContext, fromChain string, to common.Address, sentAmt *amount.Amount) error {
	flag := cc.ContractData([]byte{tagSendMaintokenFlag})
	if len(flag) == 0 {
		return nil
	}

	received := cc.AccountData(to, []byte{tagReceivedMaintoken})
	if len(received) > 0 {
		return nil
	}

	bs := cc.ContractData(makeSendMaintokenInfoKey(fromChain))
	if len(bs) == 0 {
		return nil
	}
	bf := bytes.NewBuffer(bs)
	si := &SendMaintokenInfo{}
	_, err := si.ReadFrom(bf)
	if err != nil {
		return err
	}
	if si.overthen.Cmp(sentAmt.Int) <= 0 {
		bs := cc.ContractData([]byte{tagMaintokenStore})
		if len(bs) == 0 {
			return nil
		}
		store := common.BytesToAddress(bs)
		if si.amt != nil && si.amt.Cmp(big.NewInt(0)) > 0 {
			if inf, err := cc.Exec(cc, *cc.MainToken(), "Allowance", []interface{}{store, cont.addr}); err != nil {
				return err
			} else if len(inf) == 0 {
				return errors.New("allowance return invalid")
			} else if am, ok := inf[0].(*amount.Amount); !ok {
				return errors.New("allowance return value invalid")
			} else if am.Cmp(si.amt.Int) < 0 {
				return nil
			}

			if inf, err := cc.Exec(cc, *cc.MainToken(), "BalanceOf", []interface{}{store}); err != nil {
				return err
			} else if len(inf) == 0 {
				return errors.New("balanceOf return invalid")
			} else if am, ok := inf[0].(*amount.Amount); !ok {
				return errors.New("balanceOf return value invalid")
			} else if am.Cmp(si.amt.Int) < 0 {
				return nil
			}

			err = safeTransferFrom(cc, *cc.MainToken(), store, to, si.amt)
			if err != nil {
				return err
			}
			cc.SetAccountData(to, []byte{tagReceivedMaintoken}, []byte{1})
		}
	}
	return nil
}

func unsetSendMaintoken(cc *types.ContractContext) {
	cc.SetContractData([]byte{tagSendMaintokenFlag}, nil)
	cc.SetContractData([]byte{tagMaintokenStore}, nil)
	removeSendMainTokenChains(cc)
	cc.SetContractData([]byte{tagSendChainList}, nil)
}

func removeSendMainTokenChains(cc *types.ContractContext) {
	bs := cc.ContractData([]byte{tagSendChainList})
	joindChain := string(bs)
	chains := strings.Split(joindChain, ";")
	for _, fromChain := range chains {
		cc.SetContractData(makeSendMaintokenInfoKey(fromChain), nil)
	}
}
