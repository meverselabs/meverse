package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/process/admin"
)

// RecalcHyperAdmin is recure to hyper staking amount
type RecalcHyperAdmin struct {
	Timestamp_ uint64
	From_      common.Address
	HyperAddrs []common.Address
}

// Timestamp returns the timestamp of the transaction
func (tx *RecalcHyperAdmin) Timestamp() uint64 {
	return tx.Timestamp_
}

// From returns the from address of the transaction
func (tx *RecalcHyperAdmin) From() common.Address {
	return tx.From_
}

// Fee returns the fee of the transaction
func (tx *RecalcHyperAdmin) Fee(p types.Process, loader types.LoaderWrapper) *amount.Amount {
	sp := p.(*Formulator)
	return sp.vault.GetDefaultFee(loader)
}

// Validate validates signatures of the transaction
func (tx *RecalcHyperAdmin) Validate(p types.Process, loader types.LoaderWrapper, signers []common.PublicHash) error {
	sp := p.(*Formulator)

	if tx.From() != sp.admin.AdminAddress(loader, p.Name()) {
		return admin.ErrUnauthorizedTransaction
	}

	for _, addr := range tx.HyperAddrs {
		acc, err := loader.Account(addr)
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
	}

	fromAcc, err := loader.Account(tx.From())
	if err != nil {
		return err
	}
	if err := fromAcc.Validate(loader, signers); err != nil {
		return err
	}

	if err := sp.vault.CheckFeePayable(p, loader, tx); err != nil {
		return err
	}
	return nil
}

// Execute updates the context by the transaction
func (tx *RecalcHyperAdmin) Execute(p types.Process, ctw *types.ContextWrapper, index uint16) error {
	sp := p.(*Formulator)

	return sp.vault.WithFee(p, ctw, tx, func() error {
		for _, addr := range tx.HyperAddrs {
			acc, err := ctw.Account(addr)
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

			sp := p.(*Formulator)

			StakingAmountMap, err := sp.GetStakingAmountMap(ctw, addr)
			if err != nil {
				return err
			}
			calc := amount.NewCoinAmount(0, 0)
			for _, StakingAmount := range StakingAmountMap {
				calc = calc.Add(StakingAmount)
			}
			frAcc.StakingAmount = calc
		}
		return nil
	})
}

// MarshalJSON is a marshaler function
func (tx *RecalcHyperAdmin) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"HyperAddrs":`)
	buffer.WriteString(`[`)
	for i, addr := range tx.HyperAddrs {
		if bs, err := addr.MarshalJSON(); err != nil {
			return nil, err
		} else {
			buffer.WriteString(`"`)
			buffer.Write(bs)
			buffer.WriteString(`"`)
		}
		if i < len(tx.HyperAddrs)-1 {
			buffer.WriteString(`,`)
		}
	}
	buffer.WriteString(`]`)
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
