package util

import (
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/key"
)

func Accounts() (*key.MemoryKey, common.Address, []key.Key, []common.Address, error) {
	userKeys := []key.Key{}
	users := []common.Address{}

	adminKey, err := key.NewMemoryKeyFromBytes(chainID, []byte{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return nil, ZeroAddress, nil, nil, err
	}
	admin := adminKey.PublicKey().Address()

	for i := 1; i < 11; i++ {
		pk, _ := key.NewMemoryKeyFromBytes(chainID, []byte{1, byte(i), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
		userKeys = append(userKeys, pk)
		users = append(users, pk.PublicKey().Address())
	}
	return adminKey, admin, userKeys, users, nil
}

// admin   			0xDC5b20847F43d67928F49Cd4f85D696b5A7617B5
// alice   			0xcca18d832a5C4fA1235e6c1cEa7E4645cca00395
// bob     			0x678658741a8A61B92EF6B5700397a83C92729d60
// charlie 			0x632B713ABAA2cBC9ef18B61678c5d65027a4d2f0
// eve 				0xf3dA6Ce653D680EBAcC26873d38F91aCf33C56Ac
// factory 			0xD9840365E3C375fF1B8c20fFC69EA8B1553B0C8d
// router  			0x989f4aB3c2A695Ee43c67077927c3ca12e1e6a7b
// pair    			0xc65bc0ED7dd73Aa66A8C0f7049e861D5CB7eb0B0
// swap   			0x0000000000000000000000000000000000000000
// token0  			0xa2C894E874E9A0ed61Ad4f35CBEdFA861c49c7BF
// token1  			0xe2E3a5952D2Ac311017280601F8354280394946e
// stableToken0  	0x83F875dBa87a5D2387B73222a19259A3a31a1495
// stableToken1   	0xC59B7d36115849B750b45E498bbE500F0282193A
// stableToken2 	0x5C0F85555eCAFc734BCE666D0eCe5E902A26DC4c
func WhoIs(addr common.Address) string {
	switch addr.String() {
	case "0xDC5b20847F43d67928F49Cd4f85D696b5A7617B5":
		return "admin"
	case "0xcca18d832a5C4fA1235e6c1cEa7E4645cca00395":
		return "alice"
	case "0x678658741a8A61B92EF6B5700397a83C92729d60":
		return "bob"
	case "0x632B713ABAA2cBC9ef18B61678c5d65027a4d2f0":
		return "charlie"
	case "0xf3dA6Ce653D680EBAcC26873d38F91aCf33C56Ac":
		return "eve"
	case "0xD9840365E3C375fF1B8c20fFC69EA8B1553B0C8d":
		return "factory"
	case "0x989f4aB3c2A695Ee43c67077927c3ca12e1e6a7b":
		return "router"
	case "0xc65bc0ED7dd73Aa66A8C0f7049e861D5CB7eb0B0":
		return "pair"
	case "0x0000000000000000000000000000000000000000":
		return "swap"
	case "0xa2C894E874E9A0ed61Ad4f35CBEdFA861c49c7BF":
		return "token0 in uniswap"
	case "0xe2E3a5952D2Ac311017280601F8354280394946e":
		return "token0 in uniswap"
	case "0x83F875dBa87a5D2387B73222a19259A3a31a1495":
		return "token[0] in stableseap"
	case "0xC59B7d36115849B750b45E498bbE500F0282193A":
		return "token[0] in stableseap"
	case "0x5C0F85555eCAFc734BCE666D0eCe5E902A26DC4c":
		return "token[0] in stableseap"
	}
	return "other"
}
