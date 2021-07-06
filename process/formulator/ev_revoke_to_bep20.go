package formulator

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/amount"
)

// RevokeToBEP20Event moves a ownership of utxos
type RevokeToBEP20Event struct {
	Height_      uint32
	Index_       uint16
	N_           uint16
	Heritor      common.Address
	RevokeAmount *amount.Amount
	BEP20Address string
}

// Height returns the height of the event
func (ev *RevokeToBEP20Event) Height() uint32 {
	return ev.Height_
}

// Index returns the index of the event
func (ev *RevokeToBEP20Event) Index() uint16 {
	return ev.Index_
}

// N returns the n of the event
func (ev *RevokeToBEP20Event) N() uint16 {
	return ev.N_
}

// SetN updates the n of the event
func (ev *RevokeToBEP20Event) SetN(n uint16) {
	ev.N_ = n
}

// MarshalJSON is a marshaler function
func (ev *RevokeToBEP20Event) MarshalJSON() ([]byte, error) {
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
	buffer.WriteString(`"heritor":`)
	if bs, err := ev.Heritor.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"RevokeAmount":`)
	if bs, err := ev.RevokeAmount.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`,`)
	buffer.WriteString(`"bep20_address":`)
	if bs, err := json.Marshal(ev.BEP20Address); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
