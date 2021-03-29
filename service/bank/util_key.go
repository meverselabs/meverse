package bank

import (
	"bytes"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/binutil"
)

var (
	tagSecret           = []byte{1, 0}
	tagPublicHash       = []byte{1, 1}
	tagNameAddress      = []byte{2, 0}
	tagAddressName      = []byte{2, 1}
	tagTransaction      = []byte{3, 0}
	tagTransactionList  = []byte{3, 1}
	tagTransferSendList = []byte{3, 2}
	tagTransferRecvList = []byte{3, 3}
	tagPending          = []byte{3, 4}
	tagPendingAddress   = []byte{3, 5}
	tagAccountAddress   = []byte{4, 1}
	tagAddressKeyHash   = []byte{4, 2}
	tagUnstaking        = []byte{5, 1}
)

func toSecretKey(name string) []byte {
	bs := make([]byte, 2+len(name))
	copy(bs, tagSecret)
	copy(bs[2:], []byte(name))
	return bs
}

func fromSecretKey(bs []byte) (string, error) {
	if bytes.Compare(bs[:2], tagSecret) != 0 {
		return "", ErrInvalidTag
	}
	return string(bs[2:]), nil
}

func toPublicHashKey(pubhash common.PublicHash) []byte {
	bs := make([]byte, 2+common.PublicHashSize)
	copy(bs, tagPublicHash)
	copy(bs[2:], pubhash[:])
	return bs
}

func toNameAddressKey(name string, addr common.Address) []byte {
	bs := make([]byte, 2+len(name)+common.AddressSize)
	copy(bs, tagNameAddress)
	copy(bs[2:], []byte(name))
	copy(bs[2+len(name):], addr[:])
	return bs
}

func toNameAddressPrefix(name string) []byte {
	bs := make([]byte, 2+len(name))
	copy(bs, tagNameAddress)
	copy(bs[2:], []byte(name))
	return bs
}

func toAddressNameKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagNameAddress)
	copy(bs[2:], addr[:])
	return bs
}

func fromNameAddress(bs []byte, name string) (common.Address, error) {
	if bytes.Compare(bs[:2], tagNameAddress) != 0 {
		return common.Address{}, ErrInvalidTag
	}
	if len(bs) != 2+len(name)+common.AddressSize {
		return common.Address{}, ErrInvalidNameAddress
	}
	var addr common.Address
	copy(addr[:], bs[2+len(name):])
	return addr, nil
}

func toTransferSendListKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagTransferSendList)
	copy(bs[2:], addr[:])
	return bs
}

func toTransferRecvListKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagTransferRecvList)
	copy(bs[2:], addr[:])
	return bs
}

func toTransactionListKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagTransactionList)
	copy(bs[2:], addr[:])
	return bs
}

func toPendingAddressKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagPendingAddress)
	copy(bs[2:], addr[:])
	return bs
}

func toAccountKey(pubhash common.PublicHash) []byte {
	bs := make([]byte, 2+common.PublicHashSize)
	copy(bs, tagAccountAddress)
	copy(bs[2:], pubhash[:])
	return bs
}

func toUnstakingKey(addr common.Address) []byte {
	bs := make([]byte, 2+common.AddressSize)
	copy(bs, tagUnstaking)
	copy(bs[2:], addr[:])
	return bs
}

func toUnstakingSubKey(haddr common.Address, UnstakedHeight uint32) []byte {
	bs := make([]byte, 6+common.AddressSize)
	copy(bs, tagUnstaking)
	copy(bs[2:], binutil.LittleEndian.Uint32ToBytes(UnstakedHeight))
	copy(bs[6:], haddr[:])
	return bs
}
