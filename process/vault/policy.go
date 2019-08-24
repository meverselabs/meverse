package vault

import (
	"bytes"

	"github.com/fletaio/fleta/common/amount"
)

// Policy defines a vault policy
type Policy struct {
	AccountCreationAmount *amount.Amount
}

// MarshalJSON is a marshaler function
func (pc *Policy) MarshalJSON() ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(`{`)
	buffer.WriteString(`"account_creation_amount":`)
	if bs, err := pc.AccountCreationAmount.MarshalJSON(); err != nil {
		return nil, err
	} else {
		buffer.Write(bs)
	}
	buffer.WriteString(`}`)
	return buffer.Bytes(), nil
}
