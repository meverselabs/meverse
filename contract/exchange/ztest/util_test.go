package test

import (
	"time"

	. "github.com/meverselabs/meverse/contract/exchange/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Util", func() {

	It("isContract", func() {
		beforeEachStable()
		cn, cdx, ctx, _ := initChain(genesis, admin)
		ctx, _ = Sleep(cn, ctx, nil, uint64(time.Now().UnixNano())/uint64(time.Second), aliceKey)

		cc, err := GetCC(ctx, factoryAddr, alice)
		Expect(err).To(Succeed())

		Expect(cc.IsContract(factoryAddr)).To(BeTrue())
		Expect(cc.IsContract(routerAddr)).To(BeTrue())
		Expect(cc.IsContract(swap)).To(BeTrue())

		Expect(cc.IsContract(alice)).To(BeFalse())
		Expect(cc.IsContract(bob)).To(BeFalse())

		RemoveChain(cdx)
		afterEach()
	})
})
