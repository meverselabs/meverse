package formulator

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/process/vault"
)

// RevokeToBEP20 is used to remove formulator account and send to BEP20
type RevokeToBEP20 struct {
	Timestamp_   uint64
	From_        common.Address
	Heritor      common.Address
	BEP20Address string
}

// Timestamp returns the timestamp of the transaction
func (tx *RevokeToBEP20) Timestamp() uint64 {
	return tx.Timestamp_
}

// From returns the from address of the transaction
func (tx *RevokeToBEP20) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *RevokeToBEP20) Fee(p types.Process, loader types.LoaderWrapper) *amount.Amount {
	sp := p.(*Formulator)
	return sp.vault.GetDefaultFee(loader)
}

// Validate validates signatures of the transaction
func (tx *RevokeToBEP20) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Formulator)

	if loader.TargetHeight() > 122863319 {
		return errors.New("expired tx")
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

	if err := sp.vault.CheckFeePayable(p, loader, tx); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *RevokeToBEP20) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	return sp.vault.WithFee(p, ctw, tx, func() error {
		acc, err := ctw.Account(tx.From_)
		if err != nil {
			return err
		}
		frAcc := acc.(*FormulatorAccount)
		frAcc.IsRevoked = true

		gaddr, err := ctw.AddressByName("fleta.gateway")
		if err != nil {
			return err
		}

		RevokeAmount := amount.MustParseAmount(frAcc.Amount.String())

		gf := amount.MustParseAmount("10.1")
		fam := sp.vault.Balance(ctw, tx.From())
		ham := sp.vault.Balance(ctw, tx.Heritor)
		if gf.Cmp(fam.Int) < 0 {
			if err := sp.vault.SubBalance(ctw, tx.From(), gf); err != nil {
				return err
			}
		} else if gf.Cmp(ham.Int) < 0 {
			if err := sp.vault.SubBalance(ctw, tx.Heritor, gf); err != nil {
				return err
			}
		} else {
			return vault.ErrInsufficientFee
		}
		if err := sp.vault.AddBalance(ctw, gaddr, gf); err != nil {
			return err
		}

		b := sp.vault.Balance(ctw, frAcc.Address_)
		if err := sp.vault.AddBalance(ctw, tx.Heritor, b); err != nil {
			return err
		}
		if err := sp.vault.SubBalance(ctw, frAcc.Address_, b); err != nil {
			return err
		}

		if err := sp.revokeFormulator(ctw, tx.From_, gaddr); err != nil {
			return err
		}
		ev := &RevokeToBEP20Event{
			Height_:      ctw.TargetHeight(),
			Index_:       65534,
			Heritor:      tx.Heritor,
			RevokeAmount: RevokeAmount,
			BEP20Address: tx.BEP20Address,
		}
		if err := ctw.EmitEvent(ev); err != nil {
			return err
		}

		return nil
	})
}

// MarshalJSON is a marshaler function
func (tx *RevokeToBEP20) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"heritor":`)
	if bs, err := tx.Heritor.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"bep20_address":`)
	if bs, err := json.Marshal(tx.BEP20Address); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
