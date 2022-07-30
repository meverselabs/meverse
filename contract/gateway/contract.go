package gateway

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
)

type GatewayContract struct {
	addr   common.Address
	master common.Address
}

func (cont *GatewayContract) Name() string {
	return "GatewayContract"
}

func (cont *GatewayContract) Address() common.Address {
	return cont.addr
}

func (cont *GatewayContract) Master() common.Address {
	return cont.master
}

func (cont *GatewayContract) Init(addr common.Address, master common.Address) {
	cont.addr = addr
	cont.master = master
}

func (cont *GatewayContract) OnCreate(cc *types.ContractContext, Args []byte) error {
	data := &GatewayContractConstruction{}
	if _, err := data.ReadFrom(bytes.NewReader(Args)); err != nil {
		return err
	}
	cc.SetContractData([]byte{tagTokenContractAddress}, data.TokenAddress[:])
	return nil
}

func (cont *GatewayContract) OnReward(cc *types.ContractContext, b *types.Block, CountMap map[common.Address]uint32) (map[common.Address]*amount.Amount, error) {
	return nil, nil
}

//////////////////////////////////////////////////
// Private Functions
//////////////////////////////////////////////////

//////////////////////////////////////////////////
// Public Writer Functions
//////////////////////////////////////////////////

func (cont *GatewayContract) AddPlatform(cc *types.ContractContext, Platform string, Amount *amount.Amount) error {
	isSender := cont.IsSender(cc, cc.From())

	if cc.From() != cont.Master() && !isSender {
		return errors.New("not token sender")
	}

	pf := strings.ToLower(Platform)
	bs := cc.ContractData(makePlatformKey(pf))
	if len(bs) == 1 && bs[0] == 1 {
		return errors.New("already added: " + Platform)
	}
	cc.SetContractData(makePlatformKey(pf), []byte{1})
	cc.SetContractData(makePlatformFeeKey(pf), Amount.Bytes())
	return nil
}

func (cont *GatewayContract) UpdatePlatform(cc *types.ContractContext, Platform string, Amount *amount.Amount) error {
	isSender := cont.IsSender(cc, cc.From())

	if cc.From() != cont.Master() && !isSender {
		return errors.New("not token sender")
	}

	pf := strings.ToLower(Platform)
	bs := cc.ContractData(makePlatformKey(pf))
	if !(len(bs) == 1 && bs[0] == 1) {
		return errors.New("not exist platform: " + Platform)
	}
	cc.SetContractData(makePlatformFeeKey(pf), Amount.Bytes())

	return nil
}

func (cont *GatewayContract) Transfer(cc *types.ContractContext, to common.Address, Amount *amount.Amount) error {
	isSender := cont.IsSender(cc, cc.From())
	if cc.From() != cont.Master() && !isSender {
		return errors.New("not token sender")
	}

	taddr := cont.TokenAddress(cc)
	if _, err := cc.Exec(cc, taddr, "Transfer", []interface{}{to, Amount}); err != nil {
		return err
	}

	return nil
}

func (cont *GatewayContract) TokenInRevert(cc *types.ContractContext, Platform, ercHash string, txid1, txid2 []byte, to common.Address, Amount *amount.Amount) error {
	isSender := cont.IsSender(cc, cc.From())

	if cc.From() != cont.Master() && !isSender {
		return errors.New("not token sender")
	}

	pf := strings.ToLower(Platform)
	bs := cc.ContractData(makePlatformKey(pf))
	if !(len(bs) == 1 && bs[0] == 1) {
		return errors.New("not support platform: " + Platform)
	}

	_, _, err := types.ParseTransactionIDBytes(txid1)
	if err != nil {
		return err
	}
	_, _, err = types.ParseTransactionIDBytes(txid2)
	if err != nil {
		return err
	}
	bsTxids := make([]byte, 12)
	if bytes.Compare(txid1, txid2) > 0 {
		copy(bsTxids[:6], txid1)
		copy(bsTxids[6:], txid2)
	} else {
		copy(bsTxids[:6], txid2)
		copy(bsTxids[6:], txid1)
	}

	bs = cc.ContractData(makeTokenInRevertKey(bsTxids))
	if len(bs) == 1 && bs[0] == 1 {
		return errors.New("already proccess txids: " + hex.EncodeToString(txid1) + ", " + hex.EncodeToString(txid2))
	}
	cc.SetContractData(makeTokenInRevertKey(bsTxids), []byte{1})

	ErcHash := hash.HexToHash(ercHash)
	cc.SetContractData(makeTokenInKey(ErcHash, pf), []byte{1})

	taddr := cont.TokenAddress(cc)
	if _, err := cc.Exec(cc, taddr, "TokenInRevert", []interface{}{Platform, ercHash, to, Amount}); err != nil {
		return err
	}
	return nil
}

func (cont *GatewayContract) TokenIn(cc *types.ContractContext, Platform string, ercHash string, to common.Address, Amount *amount.Amount) error {
	isSender := cont.IsSender(cc, cc.From())

	if cc.From() != cont.Master() && !isSender {
		return errors.New("not token sender")
	}

	pf := strings.ToLower(Platform)
	bs := cc.ContractData(makePlatformKey(pf))
	if !(len(bs) == 1 && bs[0] == 1) {
		return errors.New("not support platform: " + Platform)
	}

	ErcHash := hash.HexToHash(ercHash)
	bs = cc.ContractData(makeTokenInKey(ErcHash, pf))
	if len(bs) == 1 && bs[0] == 1 {
		return errors.New("exist hash: " + ercHash)
	}
	if cc.TargetHeight() > 1783888 {
		cc.SetContractData(makeTokenInKey(ErcHash, pf), []byte{1})
	}

	taddr := cont.TokenAddress(cc)
	if _, err := cc.Exec(cc, taddr, "Transfer", []interface{}{to, Amount}); err != nil {
		return err
	}
	return nil
}

func (cont *GatewayContract) TokenIndexIn(cc *types.ContractContext, Platform string, ercHash string, to common.Address, Amount *amount.Amount) error {
	err := cont.TokenIn(cc, Platform, ercHash, to, Amount)
	if err != nil {
		return err
	}
	pf := strings.ToLower(Platform)
	ErcHash := hash.HexToHash(ercHash)
	cc.SetContractData(makeTokenInKey(ErcHash, pf), []byte{1})
	return nil
}

func (cont *GatewayContract) TokenOut(cc *types.ContractContext, Platform string, withdrawAddress common.Address, Amount *amount.Amount) error {
	taddr := cont.TokenAddress(cc)

	pf := strings.ToLower(Platform)
	feebs := cc.ContractData(makePlatformFeeKey(pf))
	fee := amount.NewAmountFromBytes(feebs)

	feeOwner := cont.FeeOwner(cc)
	if feeOwner == common.ZeroAddr {
		feeOwner = cont.Master()
	}
	if _, err := cc.Exec(cc, taddr, "TransferFrom", []interface{}{cc.From(), feeOwner, fee}); err != nil {
		return err
	}
	if _, err := cc.Exec(cc, taddr, "TransferFrom", []interface{}{cc.From(), cont.Address(), Amount}); err != nil {
		return err
	}
	return nil
}

func (cont *GatewayContract) TokenLeave(cc *types.ContractContext, CoinTXID string, ERC20TXID string, Platform string) error {
	isSender := cont.IsSender(cc, cc.From())

	if cc.From() != cont.Master() && !isSender {
		return errors.New("not token sender")
	}

	pf := strings.ToLower(Platform)
	bs := cc.ContractData(makePlatformKey(pf))
	if !(len(bs) == 1 && bs[0] == 1) {
		return errors.New("not support platform: " + Platform)
	}
	return nil
}

func (cont *GatewayContract) SetSender(cc *types.ContractContext, To common.Address, Is bool) error {
	if cc.From() != cont.Master() {
		return errors.New("not token master")
	}

	isMinter := cont.IsSender(cc, To)

	if Is {
		if isMinter {
			return errors.New("already token sender")
		}
		cc.SetAccountData(To, []byte{tagTokenSender}, []byte{1})
	} else {
		if !isMinter {
			return errors.New("not token sender")
		}
		cc.SetAccountData(To, []byte{tagTokenSender}, nil)
	}
	return nil
}

func (cont *GatewayContract) SetFeeOwner(cc *types.ContractContext, feeOwner common.Address) error {
	if cc.From() != cont.Master() {
		return errors.New("not token master")
	}
	cc.SetContractData([]byte{tagFeeOwner}, feeOwner.Bytes())
	return nil
}

//////////////////////////////////////////////////
// Public Reader Functions
//////////////////////////////////////////////////

func (cont *GatewayContract) TokenAddress(cc *types.ContractContext) common.Address {
	return common.BytesToAddress(cc.ContractData([]byte{tagTokenContractAddress}))
}

func (cont *GatewayContract) FeeOwner(cc *types.ContractContext) common.Address {
	return common.BytesToAddress(cc.ContractData([]byte{tagFeeOwner}))
}

func (cont *GatewayContract) IsSender(cc types.ContractLoader, addr common.Address) bool {
	bs := cc.AccountData(addr, []byte{tagTokenSender})
	if len(bs) == 1 && bs[0] == 1 {
		return true
	}
	return false
}
