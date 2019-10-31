package bank

import (
	"bytes"

	"github.com/fletaio/fleta/common"
)

var (
	tagSecret      = []byte{1, 0}
	tagPublicHash  = []byte{1, 1}
	tagNameAddress = []byte{2, 0}
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
