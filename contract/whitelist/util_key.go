package whitelist

import "github.com/meverselabs/meverse/common/hash"

var (
	tagOwner      = byte(0x01)
	tagSeq        = byte(0x02)
	tagGroupOwner = byte(0x03)
	tagGroupData  = byte(0x04)
)

func makeWhiteListKey(key byte, body []byte) []byte {
	bs := make([]byte, 1+len(body))
	bs[0] = key
	copy(bs[1:], body[:])
	return bs
}

func makeGroupOwnerKey(id hash.Hash256) []byte {
	return makeWhiteListKey(tagGroupOwner, id.Bytes())
}
func makeGroupDataKey(id hash.Hash256) []byte {
	return makeWhiteListKey(tagGroupData, id.Bytes())
}
