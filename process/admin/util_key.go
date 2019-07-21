package admin

// tags
var (
	tagAdminAddress = []byte{1, 1}
)

func toAdminAddressKey(Name string) []byte {
	bs := make([]byte, 2+len(Name))
	copy(bs, tagAdminAddress)
	copy(bs[2:], []byte(Name))
	return bs
}
