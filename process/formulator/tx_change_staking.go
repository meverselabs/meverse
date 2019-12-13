package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
)

// ChangeStaking is used to stake coin to the hyper formulator
type ChangeStaking struct {
	Timestamp_     uint64
	Seq_           uint64
	From_          common.Address
	HyperUnstaking common.Address
	HyperStaking   common.Address
	Amount         *amount.Amount
}

// Timestamp returns the timestamp of the transaction
func (tx *ChangeStaking) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *ChangeStaking) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *ChangeStaking) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *ChangeStaking) Fee(p types.Process, loader types.LoaderWrapper) *amount.Amount {
	sp := p.(*Formulator)
	return sp.vault.GetDefaultFee(loader)
}

// Validate validates signatures of the transaction
func (tx *ChangeStaking) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Formulator)

	if tx.Amount.Less(amount.COIN.DivC(10)) {
		return ErrInvalidStakingAmount
	}
	if tx.From() == tx.HyperStaking {
		return ErrInvalidStakingAddress
	}

	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	frUnstaking, err := toHyperFormulator(loader, tx.HyperUnstaking)
	if err != nil {
		return err
	}

	fromStakingAmount := sp.GetStakingAmount(loader, tx.HyperUnstaking, tx.From())
	if fromStakingAmount.Less(tx.Amount) {
		return ErrInsufficientStakingAmount
	}
	if frUnstaking.StakingAmount.Less(tx.Amount) {
		return ErrInsufficientStakingAmount
	}

	frStaking, err := toHyperFormulator(loader, tx.HyperStaking)
	if err != nil {
		return err
	}
	if frStaking.IsRevoked {
		return ErrRevokedFormulator
	}
	if !frStaking.Policy.MinimumStaking.IsZero() && tx.Amount.Less(frStaking.Policy.MinimumStaking) {
		return ErrInvalidStakingAmount
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
func (tx *ChangeStaking) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	frUnstaking, err := toHyperFormulator(ctw, tx.HyperUnstaking)
	if err != nil {
		return err
	}
	if err := sp.subStakingAmount(ctw, tx.HyperUnstaking, tx.From(), tx.Amount); err != nil {
		return err
	}
	frUnstaking.StakingAmount = frUnstaking.StakingAmount.Sub(tx.Amount)

	frStaking, err := toHyperFormulator(ctw, tx.HyperStaking)
	if err != nil {
		return err
	}

	if err := sp.vault.CheckFeePayable(p, ctw, tx); err != nil {
		Fee := tx.Fee(p, ctw)
		if err := sp.vault.AddCollectedFee(ctw, Fee); err != nil {
			return err
		}
		sp.AddStakingAmount(ctw, tx.HyperStaking, tx.From(), tx.Amount.Sub(Fee))
		frStaking.StakingAmount = frStaking.StakingAmount.Add(tx.Amount.Sub(Fee))
		return nil
	} else {
		return sp.vault.WithFee(p, ctw, tx, func() error {
			sp.AddStakingAmount(ctw, tx.HyperStaking, tx.From(), tx.Amount)
			frStaking.StakingAmount = frStaking.StakingAmount.Add(tx.Amount)
			return nil
		})
	}
}

// MarshalJSON is a marshaler function
func (tx *ChangeStaking) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"timestamp":`)
	if bs, err := json.Marshal(tx.Timestamp_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"seq":`)
	if bs, err := json.Marshal(tx.Seq_); err != nil {
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
	buffer.WriteString(`"hyper_unstaking":`)
	if bs, err := tx.HyperUnstaking.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"hyper_staking":`)
	if bs, err := tx.HyperStaking.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"amount":`)
	if bs, err := tx.Amount.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
