package nft1155

var (
	tagName   = byte(0x01)
	tagSymbol = byte(0x02)
)

func makePoolKey(key byte, body []byte) []byte {
	bs := make([]byte, 1+len(body))
	bs[0] = key
	copy(bs[1:], body[:])
	return bs
}
