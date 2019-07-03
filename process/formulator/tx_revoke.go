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
	From       common.Address
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

// Fee returns the fee of the transaction
func (tx *Revoke) Fee(loader types.LoaderWrapper) *amount.Amount {
	return amount.COIN.DivC(10)
}

// Validate validates signatures of the transaction
func (tx *Revoke) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	if tx.Seq() <= loader.Seq(tx.From) {
		return types.ErrInvalidSequence
	}

	acc, err := loader.Account(tx.From)
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

	sn := ctw.Snapshot()
	defer ctw.Revert(sn)

	if tx.Seq() != ctw.Seq(tx.From)+1 {
		return types.ErrInvalidSequence
	}
	ctw.AddSeq(tx.From)

	heritorAcc, err := ctw.Account(tx.From)
	if err != nil {
		return err
	}

	acc, err := ctw.Account(tx.From)
	if err != nil {
		return err
	}
	frAcc, is := acc.(*FormulatorAccount)
	if !is {
		return types.ErrInvalidAccountType
	}
	switch frAcc.FormulatorType {
	case AlphaFormulatorType:
		policy := &AlphaPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData([]byte("AlphaPolicy")), &policy); err != nil {
			return err
		}
		if err := sp.vault.SubBalance(ctw, frAcc.Address(), tx.Fee(ctw)); err != nil {
			return err
		}
		sp.vault.AddLockedBalance(ctw, heritorAcc.Address(), frAcc.Amount, ctw.TargetHeight()+policy.AlphaUnlockRequiredBlocks)
		sp.vault.AddBalance(ctw, heritorAcc.Address(), sp.vault.Balance(ctw, frAcc.Address()))
	case SigmaFormulatorType:
		policy := &SigmaPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData([]byte("SigmaPolicy")), &policy); err != nil {
			return err
		}

		if err := sp.vault.SubBalance(ctw, frAcc.Address(), tx.Fee(ctw)); err != nil {
			return err
		}
		sp.vault.AddLockedBalance(ctw, heritorAcc.Address(), frAcc.Amount, ctw.TargetHeight()+policy.SigmaUnlockRequiredBlocks)
		sp.vault.AddBalance(ctw, heritorAcc.Address(), sp.vault.Balance(ctw, frAcc.Address()))
	case OmegaFormulatorType:
		policy := &OmegaPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData([]byte("OmegaPolicy")), &policy); err != nil {
			return err
		}

		if err := sp.vault.SubBalance(ctw, frAcc.Address(), tx.Fee(ctw)); err != nil {
			return err
		}
		sp.vault.AddLockedBalance(ctw, heritorAcc.Address(), frAcc.Amount, ctw.TargetHeight()+policy.OmegaUnlockRequiredBlocks)
		sp.vault.AddBalance(ctw, heritorAcc.Address(), sp.vault.Balance(ctw, frAcc.Address()))
	case HyperFormulatorType:
		policy := &HyperPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData([]byte("HyperPolicy")), &policy); err != nil {
			return err
		}

		if err := sp.vault.SubBalance(ctw, frAcc.Address(), tx.Fee(ctw)); err != nil {
			return err
		}
		sp.vault.AddLockedBalance(ctw, heritorAcc.Address(), frAcc.Amount, ctw.TargetHeight()+policy.HyperUnlockRequiredBlocks)
		sp.vault.AddBalance(ctw, heritorAcc.Address(), sp.vault.Balance(ctw, frAcc.Address()))

		keys, err := ctw.AccountDataKeys(tx.From, tagStaking)
		if err != nil {
			return err
		}
		for _, k := range keys {
			if addr, is := fromStakingKey(k); is {
				bs := ctw.AccountData(tx.From, k)
				if len(bs) == 0 {
					return ErrInvalidStakingAddress
				}
				StakingAmount := amount.NewAmountFromBytes(bs)
				if frAcc.StakingAmount.Less(StakingAmount) {
					return ErrCriticalStakingAmount
				}
				frAcc.StakingAmount.Sub(StakingAmount)

				sp.vault.AddLockedBalance(ctw, addr, StakingAmount, ctw.TargetHeight()+policy.StakingUnlockRequiredBlocks)
			}
		}
		if !frAcc.StakingAmount.IsZero() {
			return ErrCriticalStakingAmount
		}
	default:
		return types.ErrInvalidAccountType
	}
	ctw.DeleteAccount(acc)

	ctw.Commit(sn)
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
	if bs, err := tx.From.MarshalJSON(); err != nil {
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
