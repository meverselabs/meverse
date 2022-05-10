package main

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/extern/txparser"
	"github.com/meverselabs/meverse/service/txsearch"
	"golang.org/x/crypto/sha3"
)

func GenKey() (string, common.Address) {
	privKey, _ := crypto.GenerateKey()

	pub := &privKey.PublicKey

	addr := crypto.PubkeyToAddress(*pub)
	return hex.EncodeToString(privKey.D.Bytes()), addr
}

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func main() {
	// var v *big.Int
	// log.Println(v.Bytes())
	// return
	// tx := etypes.NewTransaction(1, common.Address{}, big.NewInt(100), 100, big.NewInt(10000000000), []byte{})
	// tx2 := etypes.NewTransaction(1, common.Address{}, big.NewInt(100), 10000, big.NewInt(10000000000), []byte{})
	// log.Println(tx.Hash().String(), tx2.Hash().String())
	// return
	// ChainID := big.NewInt(0x0e)

	// bs, err := rlpparser.ParseString("f8aa0d85e8d4a51000830b71b094e08fbad440dff3267f5a42061d64fc3b953c1ef180b844a9059cbb000000000000000000000000c42024ae9a4fad398322d39e7e9aab61bc5c6fe10000000000000000000000000000000000000000000000000de0b6b3a764000040a040c00054fa31fa0ccacd2e2e47c1650937bb03005ff4f56c06d059661ae10a45a06302d42ce8b4ef654ced6564068415c353b4c32c927e7e47fb500e648e01c840")
	// txinfo, err := rlpparser.ParseRLP(bs, ChainID)
	// log.Println(txinfo.Data)
	// // return
	dat, err := ioutil.ReadFile("./IERC20.json")
	fmt.Printf("var IERC20 = []byte{")
	for i, v := range dat {
		if i == len(dat)-1 {
			fmt.Printf("%v", v)
		} else {
			fmt.Printf("%v, ", v)
		}
		if i%31 == 30 {
			fmt.Printf("\n")
		}
	}
	fmt.Printf("}")

	return

	reader := bytes.NewReader(txparser.IERC20)

	// f, err := os.Open("./IERC20.json")
	// if err != nil {
	// 	panic(err)
	// }
	a, err := abi.JSON(reader)
	if err != nil {
		panic(err)
	}

	log.Println()

	rlpbs, err := hex.DecodeString("f8aa0d85e8d4a51000830b71b094e08fbad440dff3267f5a42061d64fc3b953c1ef180b844a9059cbb000000000000000000000000c42024ae9a4fad398322d39e7e9aab61bc5c6fe10000000000000000000000000000000000000000000000000de0b6b3a764000040a040c00054fa31fa0ccacd2e2e47c1650937bb03005ff4f56c06d059661ae10a45a06302d42ce8b4ef654ced6564068415c353b4c32c927e7e47fb500e648e01c840")
	if err != nil {
		panic(err)
	}
	etx, _, err := txparser.EthTxFromRLP(rlpbs)
	if err != nil {
		panic(err)
	}

	var args abi.Arguments
	data := etx.Data()[4:]
	if method, ok := a.Methods["transfer"]; ok {
		if len(data)%32 != 0 {
			panic(fmt.Errorf("abi: improperly formatted output: %s - Bytes: [%+v]", string(data), data))
		}
		args = method.Inputs
	}

	obj, err := args.Unpack(data)
	if err != nil {
		panic(err)
	}
	log.Println(obj)
	return

	var i uint32
	var j uint16
	return
	for i = 0; i < 4294967295; i += 10000 {
		for j = 0; j < 64000; j += 1000 {
			testStruct := txsearch.TxID{
				Height: i,
				Index:  j,
			}
			reqBodyBytes := new(bytes.Buffer)
			json.NewEncoder(reqBodyBytes).Encode(testStruct)
			txid := types.TransactionIDBytes(i, j)
			bf, _, _ := big.ParseFloat(fmt.Sprintf("%v.%v", i, j), 10, 48, big.ToPositiveInf)
			bs, err := bf.GobEncode()
			if err != nil {
				panic(err)
			}

			sk, addr := GenKey()
			{
				var b bytes.Buffer
				gz := gzip.NewWriter(&b)
				n, err := gz.Write(addr[:])
				log.Println(n, err)
				err = gz.Close()
				log.Println(err, len(addr), len(b.Bytes()))
			}
			{
				var b bytes.Buffer
				gz := gzip.NewWriter(&b)
				n, err := gz.Write([]byte(sk))
				log.Println(n, err)
				err = gz.Close()
				log.Println(err, len(sk), len(b.Bytes()))
			}

			log.Println(i, j, len(reqBodyBytes.Bytes()), len(txid), bf.String(), len(bs))
			time.Sleep(time.Millisecond * 10)
		}
	}
}
