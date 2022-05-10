package main

import (
	"encoding/hex"
	"log"
	"math/big"
	"strings"

	etypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/extern/txparser"
)

func main() {
	// splitNTest()
	// isZeroBigInt()
	// createAlphaRlpType2()
	findRlpSender()
}

func findRlpSender() {
	rlp := "0xf86f8085e8d4a510008307a12094c42024ae9a4fad398322d39e7e9aab61bc5c6fe1880de0b6b3a76400008082020da01b2b10a2e8f76b81b465c82f3f75553300abd24325142b9d9a1caac26cac5bbea065981cbb6bce35fcebe03b2c308a3e406e1e7d9f875a788a1692850615a7bf60"
	rlp = strings.Replace(rlp, "0x", "", 1)
	rlpBytes, err := hex.DecodeString(rlp)
	if err != nil {
		panic(err)
	}
	etx, sig, err := txparser.EthTxFromRLP(rlpBytes)
	if err != nil {
		panic(err)
	}

	log.Println(hex.EncodeToString(sig))

	// signedTx, err := types.SignTx(_tx, types.NewEIP155Signer(cfg.ChainID), privKey)

	signer := etypes.NewLondonSigner(etx.ChainId())

	log.Println(etx.ChainId().String())
	log.Println(etx.Type())
	log.Println(etx.Hash())

	pubk, err := common.RecoverPubkey(etx.ChainId(), signer.Hash(etx), sig)

	if err != nil {
		panic(err)
	}
	log.Println(pubk.Address())
}

func createAlphaRlpType2() {
	rlp := "0x02f8700e1a849502f900853ae3d043de8307a12094baa3c856fba6ffada189d6bd0a89d5ef7959c75e8084f9a29905c080a04f860709179d6a5dca4ac2ba6670c51cf48991e40a983535b71987bde39a066fa00ba2546d5634dff528f6ea40e9bd3e0d6d0ded148c3ee4622e400f41c4baacba"
	rlp = strings.Replace(rlp, "0x", "", 1)
	rlpBytes, err := hex.DecodeString(rlp)
	if err != nil {
		panic(err)
	}
	etx, sig, err := txparser.EthTxFromRLP(rlpBytes)
	if err != nil {
		panic(err)
	}

	log.Println(etx.ChainId().String(), etx, sig)
}

func isZeroBigInt() {
	bi := big.NewInt(0)
	if len(bi.Bytes()) == 0 {
		panic(0)
	}
}

func splitNTest() {
	{
		ls := strings.SplitN("method.test", ".", 2)
		if len(ls) != 2 {
			panic(1)
		}
	}
	{
		ls := strings.SplitN("method.test.test2", ".", 2)
		if len(ls) != 2 {
			panic(1)
		}
	}
}
