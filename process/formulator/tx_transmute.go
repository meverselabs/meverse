package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// Transmute is used to ustake coin from the hyper formulator
type Transmute struct {
	Timestamp_       uint64
	From_            common.Address
	Name             string
	KeyHash          common.PublicHash
	GenHash          common.PublicHash
	HyperFormulators []common.Address
	Amounts          []*amount.Amount
}

// Timestamp returns the timestamp of the transaction
func (tx *Transmute) Timestamp() uint64 {
	return tx.Timestamp_
}

// From returns the from address of the transaction
func (tx *Transmute) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *Transmute) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Formulator)

	if !types.IsAllowedAccountName(tx.Name) {
		return types.ErrInvalidAccountName
	}

	transmutePolicy, err := sp.GetTransmutePolicy(loader)
	if err != nil {
		return err
	}
	if loader.TargetHeight() < transmutePolicy.TransmuteEnableHeightFrom {
		return ErrInvalidTransmuteHeight
	}
	if loader.TargetHeight() > transmutePolicy.TransmuteEnableHeightTo {
		return ErrInvalidTransmuteHeight
	}

	if len(tx.Amounts) > 10 {
		return ErrInvalidTransmuteCount
	}
	if len(tx.Amounts) != len(tx.HyperFormulators) {
		return ErrInvalidTransmuteCount
	}

	if has, err := loader.HasAccountName(tx.Name); err != nil {
		return err
	} else if has {
		return types.ErrExistAccountName
	}

	alphaPolicy := &AlphaPolicy{}
	if err := encoding.Unmarshal(loader.ProcessData(tagAlphaPolicy), &alphaPolicy); err != nil {
		return err
	}
	if loader.TargetHeight() < alphaPolicy.AlphaCreationLimitHeight {
		return ErrAlphaCreationLimited
	}

	sum := amount.NewCoinAmount(0, 0)
	for _, am := range tx.Amounts {
		if am.Less(amount.COIN) {
			return ErrInvalidStakingAmount
		}
		sum = sum.Add(am)
	}
	if !sum.Equal(alphaPolicy.AlphaCreationAmount) {
		return ErrInvalidTransmuteAmount
	}

	for i, haddr := range tx.HyperFormulators {
		am := tx.Amounts[i]

		acc, err := loader.Account(haddr)
		if err != nil {
			return err
		}
		frAcc, is := acc.(*FormulatorAccount)
		if !is {
			return types.ErrInvalidAccountType
		}
		if frAcc.FormulatorType != HyperFormulatorType {
			return types.ErrInvalidAccountType
		}

		fromStakingAmount := sp.GetStakingAmount(loader, haddr, tx.From())
		if fromStakingAmount.Less(am) {
			return ErrInsufficientStakingAmount
		}
		if frAcc.StakingAmount.Less(am) {
			return ErrInsufficientStakingAmount
		}
	}

	fromAcc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	if err := fromAcc.Validate(loader, signers); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *Transmute) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	transmutePolicy, err := sp.GetTransmutePolicy(ctw)
	if err != nil {
		return err
	}
	if ctw.TargetHeight() < transmutePolicy.TransmuteEnableHeightFrom {
		return ErrInvalidTransmuteHeight
	}
	if ctw.TargetHeight() > transmutePolicy.TransmuteEnableHeightTo {
		return ErrInvalidTransmuteHeight
	}

	for i, haddr := range tx.HyperFormulators {
		am := tx.Amounts[i]

		acc, err := ctw.Account(haddr)
		if err != nil {
			return err
		}
		frAcc := acc.(*FormulatorAccount)

		if err := sp.subStakingAmount(ctw, haddr, tx.From(), am); err != nil {
			return err
		}
		frAcc.StakingAmount = frAcc.StakingAmount.Sub(am)
	}

	policy := &AlphaPolicy{}
	if err := encoding.Unmarshal(ctw.ProcessData(tagAlphaPolicy), &policy); err != nil {
		return err
	}

	acc := &FormulatorAccount{
		Address_:       sp.cn.NewAddress(ctw.TargetHeight(), index),
		Name_:          tx.Name,
		FormulatorType: AlphaFormulatorType,
		KeyHash:        tx.KeyHash,
		GenHash:        tx.GenHash,
		Amount:         policy.AlphaCreationAmount,
		PreHeight:      0,
		UpdatedHeight:  ctw.TargetHeight(),
		RewardCount:    0,
	}
	if err := ctw.CreateAccount(acc); err != nil {
		return err
	}
	return nil
}

// MarshalJSON is a marshaler function
func (tx *Transmute) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"timestamp":`)
	if bs, err := json.Marshal(tx.Timestamp_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"from":`)
	if bs, err := tx.From_.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"name":`)
	if bs, err := json.Marshal(tx.Name); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"key_hash":`)
	if bs, err := tx.KeyHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"gen_hash":`)
	if bs, err := tx.GenHash.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"hyper_formulators":`)
	buffer.WriteString(`[`)
	for i, To := range tx.HyperFormulators {
		if i > 0 {
			buffer.WriteString(`,`)
		}
		if bs, err := To.MarshalJSON(); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
	}
	buffer.WriteString(`]`)
	buffer.WriteString(`,`)
	buffer.WriteString(`"amounts":`)
	buffer.WriteString(`[`)
	for i, To := range tx.Amounts {
		if i > 0 {
			buffer.WriteString(`,`)
		}
		if bs, err := To.MarshalJSON(); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
	}
	buffer.WriteString(`]`)
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
