package engin

import "github.com/meverselabs/meverse/common/bin"

var (
	tagEngin        = byte(0x01)
	tagContractSeq  = byte(0x02)
	tagDescription  = byte(0x03)
	tagEnginVersion = byte(0x04)
)

func makeEnginURLKey(name string, version uint32) []byte {
	n := []byte(name)
	bs := make([]byte, 1+4+len(n))
	bs[0] = tagEngin
	copy(bs[1:], bin.Uint32Bytes(version))
	copy(bs[5:], n)
	return bs
}

func makeEnginVersionKey(name string) []byte {
	n := []byte(name)
	bs := make([]byte, 1+len(n))
	bs[0] = tagEnginVersion
	copy(bs[1:], n)
	return bs
}

func makeDescriptionKey(name string, version uint32) []byte {
	n := []byte(name)
	bs := make([]byte, 1+4+len(n))
	bs[0] = tagDescription
	copy(bs[1:], bin.Uint32Bytes(version))
	copy(bs[5:], n)
	return bs
}
