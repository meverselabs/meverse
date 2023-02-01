package test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// test : ginkgo
//        ginkgo -v  (verbose mode)
// skip : It("...", func() {
//           if(condition)  {
//			 	Skip("생략이유")
//           }
//         })
// focus : It -> FIt,  Describe -> FDescribe

// uniswap
// files :
// 	v2-core/test/UniswapV2ERC20.spec.ts
// 	v2-core/test/UniswapV2Factory.spec.ts
// 	v2-core/test/UniswapV2Pair.spec.ts
// 	v2-perphery/test/UniswapV2Router01.spec.ts

// stable : curve-contract
// files :
// 	tests/forked/test_gas.py
// 	tests/forked/test_insufficient_balances.py
// 	tests/pools/common/integration/test_curve.py
// 	tests/pools/common/integration/test_heavily_imbalanced.py
// 	tests/pools/common/integration/test_virtual_price_increases.py
// 	tests/pools/common/unitary/test_add_liquidity.py
// 	tests/pools/common/unitary/test_add_liquidity_initial.py
// 	tests/pools/common/unitary/test_claim_fees.py
// 	tests/pools/common/unitary/test_exchange.py
// 	tests/pools/common/unitary/test_exchange_reverts.py
// 	tests/pools/common/unitary/test_get_virtual_price.py
// 	tests/pools/common/unitary/test_kill.py
// 	tests/pools/common/unitary/test_modify_fees.py
// 	tests/pools/common/unitary/test_nonpayable.py
// 	tests/pools/common/unitary/test_ramp_A_precise.py
// 	tests/pools/common/unitary/test_remove_liquidity.py
// 	tests/pools/common/unitary/test_remove_liquidity_imbalance.py
// 	tests/pools/common/unitary/test_remove_liquidity_one_coin.py
// 	tests/pools/common/unitary/test_transfer_ownership.py
// 	tests/pools/common/unitary/test_xfer_to_contract.py

// 	tests/zaps  : skip

func TestExchange(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ERC20WrapperTest Suite")
}
