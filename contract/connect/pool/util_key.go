package pool

var (
	// tagOwner = byte(0x01)
	tagGov  = byte(0x02)
	tagFarm = byte(0x03)
	tagWant = byte(0x04)

	tagFeeFundAddress    = byte(0x07)
	tagRewardsAddress    = byte(0x08)
	tagDepositFeeFactor  = byte(0x09)
	tagWithdrawFeeFactor = byte(0x10)
	tagRewardFeeFactor   = byte(0x11)

	tagLastEarnBlock   = byte(0x12)
	tagWantLockedTotal = byte(0x13)
	tagSharesTotal     = byte(0x14)

	tagHoldShares       = byte(0x15)
	tagHoldSharesHeight = byte(0x16)
)

func makeFarmKey(key byte, body []byte) []byte {
	bs := make([]byte, 1+len(body))
	bs[0] = key
	copy(bs[1:], body[:])
	return bs
}

// func makeUserInfoKey(pid uint64, user common.Address) []byte {
// 	bs := append(bin.Uint64Bytes(pid), user[:]...)
// 	return makeFarmKey(tagUserInfo, bs)
// }
