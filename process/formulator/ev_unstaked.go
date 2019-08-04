package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
)

// UnstakedEvent moves a ownership of utxos
type UnstakedEvent struct {
	Height_         uint32
	Index_          uint16
	N_              uint16
	HyperFormulator common.Address
	Address         common.Address
	Amount          *amount.Amount
}

// Height returns the height of the event
func (ev *UnstakedEvent) Height() uint32 {
	return ev.Height_
}

// Index returns the index of the event
func (ev *UnstakedEvent) Index() uint16 {
	return ev.Index_
}

// N returns the n of the event
func (ev *UnstakedEvent) N() uint16 {
	return ev.N_
}

// SetN updates the n of the event
func (ev *UnstakedEvent) SetN(n uint16) {
	ev.N_ = n
}

// MarshalJSON is a marshaler function
func (ev *UnstakedEvent) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"height":`)
	if bs, err := json.Marshal(ev.Height_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"index":`)
	if bs, err := json.Marshal(ev.Index_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"n":`)
	if bs, err := json.Marshal(ev.N_); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"hyper_formulator":`)
	if bs, err := ev.HyperFormulator.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"address":`)
	if bs, err := ev.Address.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"amount":`)
	if bs, err := ev.Amount.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
