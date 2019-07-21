package gateway

import (
	"bytes"
	"encoding/json"

	"github.com/fletaio/fleta/common/amount"
)

// Policy defines a policy of gateway
type Policy struct {
	WithdrawFee *amount.Amount
}

// MarshalJSON is a marshaler function
func (pc *Policy) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"withdraw_fee":`)
	if bs, err := json.Marshal(pc.WithdrawFee); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
