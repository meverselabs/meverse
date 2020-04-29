package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// CreateSigma is used to make sigma formulator account
type CreateSigma struct {
	Timestamp_       uint64
	Seq_             uint64
	From_            common.Address
	AlphaFormulators []common.Address
}

// Timestamp returns the timestamp of the transaction
func (tx *CreateSigma) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *CreateSigma) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *CreateSigma) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *CreateSigma) Fee(p types.Process, loader types.LoaderWrapper) *amount.Amount {
	sp := p.(*Formulator)
	return sp.vault.GetDefaultFee(loader)
}

// Validate validates signatures of the transaction
func (tx *CreateSigma) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Formulator)

	if tx.From() != tx.AlphaFormulators[0] {
		return ErrInvalidFormulatorAddress
	}
	if tx.Seq() <= loader.Seq(tx.AlphaFormulators[0]) {
		return types.ErrInvalidSequence
	}
	if len(tx.AlphaFormulators) != len(signers) {
		return types.ErrInvalidSignerCount
	}

	policy := &SigmaPolicy{}
	if err := encoding.Unmarshal(loader.ProcessData(tagSigmaPolicy), &policy); err != nil {
		return err
	}
	if policy.SigmaRequiredAlphaCount == 0 {
		return ErrSigmaCreationNotAllowed
	}
	if len(tx.AlphaFormulators) != int(policy.SigmaRequiredAlphaCount) {
		return ErrInvalidFormulatorCount
	}

	rewardPolicy, err := sp.GetRewardPolicy(loader)
	if err != nil {
		return err
	}

	for i, From := range tx.AlphaFormulators {
		acc, err := loader.Account(From)
		if err != nil {
			return err
		}
		frAcc, is := acc.(*FormulatorAccount)
		if !is {
			return types.ErrInvalidAccountType
		}
		if frAcc.FormulatorType != AlphaFormulatorType {
			return types.ErrInvalidAccountType
		}
		if loader.TargetHeight()+frAcc.PreHeight < frAcc.UpdatedHeight+policy.SigmaRequiredAlphaBlocks {
			return ErrInsufficientFormulatorBlocks
		}
		if err := frAcc.Validate(loader, []common.PublicHash{signers[i]}); err != nil {
			return err
		}
		if sp.IsRewardBaseUpgrade(loader) {
			if frAcc.RewardCount*rewardPolicy.PayRewardEveryBlocks+frAcc.PreHeight < policy.SigmaRequiredAlphaBlocks {
				return ErrInsufficientFormulatorRewardCount
			}
		}
	}

	if err := sp.vault.CheckFeePayable(p, loader, tx); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *CreateSigma) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	return sp.vault.WithFee(p, ctw, tx, func() error {
		policy := &SigmaPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData(tagSigmaPolicy), &policy); err != nil {
			return err
		}
		acc, err := ctw.Account(tx.AlphaFormulators[0])
		if err != nil {
			return err
		}
		frAcc := acc.(*FormulatorAccount)
		for _, addr := range tx.AlphaFormulators[1:] {
			if acc, err := ctw.Account(addr); err != nil {
				return err
			} else {
				subAcc := acc.(*FormulatorAccount)
				if err := sp.vault.AddBalance(ctw, tx.AlphaFormulators[0], sp.vault.Balance(ctw, addr)); err != nil {
					return err
				}
				if err := sp.vault.RemoveBalance(ctw, addr); err != nil {
					return err
				}
				frAcc.Amount = frAcc.Amount.Add(subAcc.Amount)
				if err := ctw.DeleteAccount(subAcc); err != nil {
					return err
				}
			}
		}

		frAcc.FormulatorType = SigmaFormulatorType
		frAcc.PreHeight = 0
		frAcc.UpdatedHeight = ctw.TargetHeight()
		frAcc.RewardCount = 0
		return nil
	})
}

// MarshalJSON is a marshaler function
func (tx *CreateSigma) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"alpha_formulators":[`)
	for i, addr := range tx.AlphaFormulators {
		if i > 0 {
			buffer.WriteString(`,`)
		}
		if bs, err := addr.MarshalJSON(); err != nil {
			return nil, err
		} else {
			buffer.Write(bs)
		}
	}
	buffer.WriteString(`]`)
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
