package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// WithdrawOverAmount is used to remove formulator account and get back staked coin
type WithdrawOverAmount struct {
	Timestamp_ uint64
	Seq_       uint64
	From_      common.Address
}

// Timestamp returns the timestamp of the transaction
func (tx *WithdrawOverAmount) Timestamp() uint64 {
	return tx.Timestamp_
}

// Seq returns the sequence of the transaction
func (tx *WithdrawOverAmount) Seq() uint64 {
	return tx.Seq_
}

// From returns the from address of the transaction
func (tx *WithdrawOverAmount) From() common.Address {
	return tx.From_
}

// Validate validates signatures of the transaction
func (tx *WithdrawOverAmount) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	if tx.Seq() <= loader.Seq(tx.From()) {
		return types.ErrInvalidSequence
	}

	acc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	frAcc, is := acc.(*FormulatorAccount)
	if !is {
		return types.ErrInvalidAccountType
	}
	if frAcc.IsRevoked {
		return ErrRevokedFormulator
	}
	if err := frAcc.Validate(loader, signers); err != nil {
		return err
	}

	switch frAcc.FormulatorType {
	case AlphaFormulatorType:
		alphaPolicy := &AlphaPolicy{}
		if err := encoding.Unmarshal(loader.ProcessData(tagAlphaPolicy), &alphaPolicy); err != nil {
			return err
		}
		requried := alphaPolicy.AlphaCreationAmount
		if !requried.Less(frAcc.Amount) {
			return ErrNoOverAmount
		}
	case SigmaFormulatorType:
		alphaPolicy := &AlphaPolicy{}
		if err := encoding.Unmarshal(loader.ProcessData(tagAlphaPolicy), &alphaPolicy); err != nil {
			return err
		}
		sigmaPolicy := &SigmaPolicy{}
		if err := encoding.Unmarshal(loader.ProcessData(tagSigmaPolicy), &sigmaPolicy); err != nil {
			return err
		}
		requried := alphaPolicy.AlphaCreationAmount.MulC(int64(sigmaPolicy.SigmaRequiredAlphaCount))
		if !requried.Less(frAcc.Amount) {
			return ErrNoOverAmount
		}
	case OmegaFormulatorType:
		alphaPolicy := &AlphaPolicy{}
		if err := encoding.Unmarshal(loader.ProcessData(tagAlphaPolicy), &alphaPolicy); err != nil {
			return err
		}
		sigmaPolicy := &SigmaPolicy{}
		if err := encoding.Unmarshal(loader.ProcessData(tagSigmaPolicy), &sigmaPolicy); err != nil {
			return err
		}
		omegaPolicy := &OmegaPolicy{}
		if err := encoding.Unmarshal(loader.ProcessData(tagOmegaPolicy), &omegaPolicy); err != nil {
			return err
		}
		requried := alphaPolicy.AlphaCreationAmount.MulC(int64(sigmaPolicy.SigmaRequiredAlphaCount)).MulC(int64(omegaPolicy.OmegaRequiredSigmaCount))
		if !requried.Less(frAcc.Amount) {
			return ErrNoOverAmount
		}
	case HyperFormulatorType:
		policy := &HyperPolicy{}
		if err := encoding.Unmarshal(loader.ProcessData(tagHyperPolicy), &policy); err != nil {
			return err
		}
		requried := policy.HyperCreationAmount
		if !requried.Less(frAcc.Amount) {
			return ErrNoOverAmount
		}
	default:
		return types.ErrInvalidAccountType
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *WithdrawOverAmount) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

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
		requried := policy.AlphaCreationAmount
		if !requried.Less(frAcc.Amount) {
			return ErrNoOverAmount
		}
		sp.vault.AddBalance(ctw, tx.From(), frAcc.Amount.Sub(requried))
		frAcc.Amount = requried
	case SigmaFormulatorType:
		alphaPolicy := &AlphaPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData(tagAlphaPolicy), &alphaPolicy); err != nil {
			return err
		}
		policy := &SigmaPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData(tagSigmaPolicy), &policy); err != nil {
			return err
		}
		requried := alphaPolicy.AlphaCreationAmount.MulC(int64(policy.SigmaRequiredAlphaCount))
		if !requried.Less(frAcc.Amount) {
			return ErrNoOverAmount
		}
		sp.vault.AddBalance(ctw, tx.From(), frAcc.Amount.Sub(requried))
		frAcc.Amount = requried
	case OmegaFormulatorType:
		alphaPolicy := &AlphaPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData(tagAlphaPolicy), &alphaPolicy); err != nil {
			return err
		}
		sigmaPolicy := &SigmaPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData(tagSigmaPolicy), &sigmaPolicy); err != nil {
			return err
		}
		policy := &OmegaPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData(tagOmegaPolicy), &policy); err != nil {
			return err
		}
		requried := alphaPolicy.AlphaCreationAmount.MulC(int64(sigmaPolicy.SigmaRequiredAlphaCount)).MulC(int64(policy.OmegaRequiredSigmaCount))
		if !requried.Less(frAcc.Amount) {
			return ErrNoOverAmount
		}
		sp.vault.AddBalance(ctw, tx.From(), frAcc.Amount.Sub(requried))
		frAcc.Amount = requried
	case HyperFormulatorType:
		policy := &HyperPolicy{}
		if err := encoding.Unmarshal(ctw.ProcessData(tagHyperPolicy), &policy); err != nil {
			return err
		}
		requried := policy.HyperCreationAmount
		if !requried.Less(frAcc.Amount) {
			return ErrNoOverAmount
		}
		sp.vault.AddBalance(ctw, tx.From(), frAcc.Amount.Sub(requried))
		frAcc.Amount = requried
	default:
		return types.ErrInvalidAccountType
	}
	return nil
}

// MarshalJSON is a marshaler function
func (tx *WithdrawOverAmount) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
