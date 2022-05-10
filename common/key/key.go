package key

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/hash"
)

// Key defines crypto key functions
type Key interface {
	Sign(h hash.Hash256) (common.Signature, error)
	SignWithPassphrase(h hash.Hash256, passphrase []byte) (common.Signature, error)
	Verify(h hash.Hash256, sig common.Signature) bool
	PublicKey() common.PublicKey
	Clear()
}
