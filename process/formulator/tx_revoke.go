package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// Revoke is used to remove formulator account and get back staked coin
type Revoke struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
	Heritor    common.Address
}

// Timestamp returns the timestamp of the transaction
func (tx *Revoke) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *Revoke) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *Revoke) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *Revoke) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	if has, err := loader.HasAccount(tx.Heritor); err != nil {
		return err
	} else if !has {
		return types.ErrNotExistAccount
	}

	acc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	frAcc, is := acc.(*FormulatorAccount)
	if !is {
		return types.ErrInvalidAccountType
	}
	if err := frAcc.Validate(loader, signers); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *Revoke) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	heritorAcc, err := ctw.Account(tx.From())
	if err != nil {
		return err
	}

	acc, err := ctw.Account(tx.From())
	if err != nil {
		return err
	}
	frAcc := acc.(*FormulatorAccount)
	switch frAcc.FormulatorType {
	case AlphaFormulatorType:
		policy := &AlphaPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData(tagAlphaPolicy), &policy); err != nil {
			return err
		}
		if err := sp.vault.SubBalance(ctw, frAcc.Address(), amount.COIN.DivC(10)); err != nil {
			return err
		}
		if err := sp.vault.AddLockedBalance(ctw, heritorAcc.Address(), ctw.TargetHeight()+policy.AlphaUnlockRequiredBlocks, frAcc.Amount); err != nil {
			return err
		}
		if err := sp.vault.AddBalance(ctw, heritorAcc.Address(), sp.vault.Balance(ctw, frAcc.Address())); err != nil {
			return err
		}
	case SigmaFormulatorType:
		policy := &SigmaPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData(tagSigmaPolicy), &policy); err != nil {
			return err
		}

		if err := sp.vault.SubBalance(ctw, frAcc.Address(), amount.COIN.DivC(10)); err != nil {
			return err
		}
		if err := sp.vault.AddLockedBalance(ctw, heritorAcc.Address(), ctw.TargetHeight()+policy.SigmaUnlockRequiredBlocks, frAcc.Amount); err != nil {
			return err
		}
		if err := sp.vault.AddBalance(ctw, heritorAcc.Address(), sp.vault.Balance(ctw, frAcc.Address())); err != nil {
			return err
		}
	case OmegaFormulatorType:
		policy := &OmegaPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData(tagOmegaPolicy), &policy); err != nil {
			return err
		}

		if err := sp.vault.SubBalance(ctw, frAcc.Address(), amount.COIN.DivC(10)); err != nil {
			return err
		}
		if err := sp.vault.AddLockedBalance(ctw, heritorAcc.Address(), ctw.TargetHeight()+policy.OmegaUnlockRequiredBlocks, frAcc.Amount); err != nil {
			return err
		}
		if err := sp.vault.AddBalance(ctw, heritorAcc.Address(), sp.vault.Balance(ctw, frAcc.Address())); err != nil {
			return err
		}
	case HyperFormulatorType:
		policy := &HyperPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData(tagHyperPolicy), &policy); err != nil {
			return err
		}

		if err := sp.vault.SubBalance(ctw, frAcc.Address(), amount.COIN.DivC(10)); err != nil {
			return err
		}
		if err := sp.vault.AddLockedBalance(ctw, heritorAcc.Address(), ctw.TargetHeight()+policy.HyperUnlockRequiredBlocks, frAcc.Amount); err != nil {
			return err
		}
		if err := sp.vault.AddBalance(ctw, heritorAcc.Address(), sp.vault.Balance(ctw, frAcc.Address())); err != nil {
			return err
		}

		PowerMap, err := sp.GetStakingAmountMap(ctw, tx.From())
		if err != nil {
			return err
		}
		for addr, StakingAmount := range PowerMap {
			if StakingAmount.IsZero() {
				return ErrInvalidStakingAddress
			}
			if frAcc.StakingAmount.Less(StakingAmount) {
				return ErrCriticalStakingAmount
			}
			frAcc.StakingAmount = frAcc.StakingAmount.Sub(StakingAmount)

			if err := sp.vault.AddLockedBalance(ctw, addr, ctw.TargetHeight()+policy.StakingUnlockRequiredBlocks, StakingAmount); err != nil {
				return err
			}
		}
		if !frAcc.StakingAmount.IsZero() {
			return ErrCriticalStakingAmount
		}
	default:
		return types.ErrInvalidAccountType
	}
	if err := ctw.DeleteAccount(acc); err != nil {
		return err
	}
	return nil
}

// MarshalJSON is a marshaler function
func (tx *Revoke) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"heritor":`)
	if bs, err := tx.Heritor.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
