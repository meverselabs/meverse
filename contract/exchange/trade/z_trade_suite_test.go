package trade

import (
	"math/big"
	"testing"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/common/key"
	"github.com/meverselabs/meverse/core/types"

	. "github.com/meverselabs/meverse/contract/exchange/util"
	"github.com/meverselabs/meverse/contract/token"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestExchange(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Trade Suite")
}

var (
	adminKey   key.Key
	admin      common.Address
	users      []common.Address
	alice, bob common.Address

	chainID  = big.NewInt(1)
	classMap = map[string]uint64{}
	genesis  *types.Context
)

var _ = BeforeSuite(func() {

	classID, _ := types.RegisterContractType(&StableSwap{})
	classMap["StableSwap"] = classID

	adminKey, admin, _, users, _ = Accounts()

	alice = admin
	bob = users[0]
})

func cleanUp() {
	genesis = nil
}

//Contract Address
func contractAddress(sender common.Address, ClassID uint64, seq uint32) common.Address {
	base := make([]byte, 1+common.AddressLength+8+4)
	base[0] = 0xff
	copy(base[1:], sender[:])
	copy(base[1+common.AddressLength:], bin.Uint64Bytes(ClassID))
	copy(base[1+common.AddressLength+8:], bin.Uint32Bytes(seq))
	h := hash.Hash(base)
	addr := common.BytesToAddress(h[12:])
	return addr
}

//Token Contract Address
func tokenContractAddress(sender common.Address, seq uint32) (common.Address, error) {
	classID, err := types.RegisterContractType(&token.TokenContract{})
	if err != nil {
		return ZeroAddress, err
	}
	return contractAddress(sender, classID, seq), nil
}
