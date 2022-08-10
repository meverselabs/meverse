package bridge

import (
	"github.com/meverselabs/meverse/common"
)

var (
	tagRouterContractAddress = byte(0x01)
	tagBankAddress           = byte(0x02)
	tagFeeOwnerAddress       = byte(0x03)
	tagMeverseTokenAddress   = byte(0x04)

	tagSequenceFrom = byte(0x05)
	tagSequenceTo   = byte(0x06)

	tagTransferFeeInfoToChain = byte(0x07)
	tagTokenFeeInfoFromChain  = byte(0x08)

	tagSendMaintokenFlag = byte(0x09)
	tagSendMaintokenInfo = byte(0x10)
	tagMaintokenStore    = byte(0x11)
	tagReceivedMaintoken = byte(0x12)
	tagSendChainList     = byte(0x13)

	tagTransferTokenFeeInfoToChain = byte(0x14)

	tagTokenFeeOwnerAddress = byte(0x15)
)

func makeBridgeKey(key byte, body []byte) []byte {
	bs := make([]byte, 1+len(body))
	bs[0] = key
	copy(bs[1:], body[:])
	return bs
}
func makeSequenceFrom(addr common.Address, toChain string) []byte {
	return makeBridgeKey(tagSequenceFrom, append(addr[:], []byte(toChain)...))
}
func makeSequenceTo(addr common.Address, fromChain string) []byte {
	return makeBridgeKey(tagSequenceTo, append(addr[:], []byte(fromChain)...))
}
func makeTransferFeeInfoToChain(chain string) []byte {
	return makeBridgeKey(tagTransferFeeInfoToChain, []byte(chain))
}
func makeTransferTokenFeeInfoToChain(chain string) []byte {
	return makeBridgeKey(tagTransferTokenFeeInfoToChain, []byte(chain))
}
func makeTokenFeeInfoFromChain(chain string) []byte {
	return makeBridgeKey(tagTokenFeeInfoFromChain, []byte(chain))
}
func makeSendMaintokenInfoKey(fromChain string) []byte {
	return makeBridgeKey(tagSendMaintokenInfo, []byte(fromChain))
}

// func makeSequenceTo(txid []byte, Platform string) []byte {
// 	bs := append(txid, []byte(Platform)...)
// 	return makeBridgeKey(tagTokenOut, bs)
// }
