package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

var fileName = "Reveal"

var read = `
`

var write = `
reveal(strs []string) string {
`

var reg = []string{
	`^\n`,
	``,
	`\n$`,
	``,
	`(cc types.ContractLoader|cc \*types.ContractContext)[, ]*`,
	``,
	`[, ]+error`,
	``,
	`(?:([a-zA-Z0-9_]+) (common.Address))`,
	`{"internalType": "address","name": "$1","type": "address"}`,
	`(?:([a-zA-Z0-9_]+) (hash.Hash256))`,
	`{"internalType": "uint256","name": "$1","type": "uint256"}`,
	`(?:([a-zA-Z0-9_]+) (\[\]common.Address))`,
	`{"internalType": "address[]","name": "$1","type": "address[]"}`,
	`(?:([a-zA-Z0-9_]+) (\[\]hash.Hash256))`,
	`{"internalType": "uint256[]","name": "$1","type": "uint256[]"}`,
	`(?:([a-zA-Z0-9_]+) (\[\]string))`,
	`{"internalType": "tokenZtring[]","name": "$1","type": "tokenZtring[]"}`,
	`(?:([a-zA-Z0-9_]+) (string))`,
	`{"internalType": "tokenZtring","name": "$1","type": "tokenZtring"}`,
	`(?:([a-zA-Z0-9_]+) (\*amount.Amount|\*big.Int|uint[0-9]+))`,
	`{"internalType": "uint256","name": "$1","type": "uint256"}`,
	`(?:([a-zA-Z0-9_]+) (\[\]\*amount.Amount|\[\]\*big.Int|\[\]uint[0-9]+))`,
	`{"internalType": "uint256[]","name": "$1","type": "uint256[]"}`,
	`(?:([a-zA-Z0-9_]+) (\[\]byte))`,
	`{"internalType": "bytes32","name": "$1","type": "bytes32"}`,
	`(?:([a-zA-Z0-9_]+) (bool))`,
	`{"internalType": "tokenBool","name": "$1","type": "tokenBool"}`,
	`(?:(\[\]common.Address))`,
	`{"internalType": "address[]","name": "","type": "address[]"}`,
	`(?:(\[\]hash.Hash256))`,
	`{"internalType": "uint256[]","name": "","type": "uint256[]"}`,
	`(?:(hash.Hash256))`,
	`{"internalType": "uint256","name": "","type": "uint256"}`,
	`(?:(common.Address))`,
	`{"internalType": "address","name": "","type": "address"}`,
	`(?:(\[\]string))`,
	` {"internalType": "string[]","name": "","type": "string[]"}`,
	`(?:( string))`,
	` {"internalType": "string","name": "","type": "string"}`,
	`(?:(\[\]\*amount.Amount|\[\]\*big.Int|\[\]uint8|\[\]uint16|\[\]uint32|\[\]uint64))`,
	`{"internalType": "uint256[]","name": "","type": "uint256[]"}`,
	`(?:(\*amount.Amount|\*big.Int|uint8|uint16|uint32|uint64))`,
	`{"internalType": "uint256","name": "","type": "uint256"}`,
	`(?:(\[\]byte))`,
	`{"internalType": "bytes32","name": "","type": "bytes32"}`,
	`(?:(bool))`,
	`{"internalType": "bool","name": "","type": "bool"}`,
	`(?:(tokenZtring\[\]))`,
	`string[]`,
	`(?:(tokenZtring))`,
	`string`,
	`(?:(tokenBool))`,
	`bool`,
}

func main() {
	// t := imo.ImoContract{}
	// f := t.Front()

	reads := strings.Split(read, "\n")
	writes := strings.Split(write, "\n")
	for i := 0; i < len(reg); i += 2 {
		// log.Println(reg[i], "**********", reg[i+1])
		m1 := regexp.MustCompile(reg[i])
		// read = m1.ReplaceAllString(read, reg[i+1])
		for j, s := range reads {
			reads[j] = m1.ReplaceAllString(s, reg[i+1])
		}
		// write = m1.ReplaceAllString(write, reg[i+1])
		for j, s := range writes {
			writes[j] = m1.ReplaceAllString(s, reg[i+1])
		}
	}

	m1 := regexp.MustCompile(`([a-zA-Z_0-9]+)[ ]*\([ ]*([\{\}\[\]$ 0-9a-zA-Z_":\t,]*)\) [(]*([^)]*)[)]*[\) ]*\{$`)

	readapi := `{"constant": true,"inputs": [$2],"name": "$1","outputs": [$3],"payable": false,"stateMutability": "view","type": "function"}`
	writeapi := `{"inputs": [$2], "name": "$1","outputs": [$3],"stateMutability": "nonpayable","type": "function"}`

	// read = m1.ReplaceAllString(read, readapi)
	for i, s := range reads {
		reads[i] = m1.ReplaceAllString(s, readapi)
		log.Println(reads[i])
	}
	// write = m1.ReplaceAllString(write, writeapi)
	for i, s := range writes {
		writes[i] = m1.ReplaceAllString(s, writeapi)
		log.Println(writes[i])
	}

	abi := append(writes, reads...)
	log.Println(abi)
	ms := []map[string]interface{}{}
	for _, v := range abi {
		m := map[string]interface{}{}
		if v == "" {
			continue
		}
		err := json.Unmarshal([]byte(v), &m)
		if err != nil {
			log.Println(v)
			panic(err)
		}
		ms = append(ms, m)
	}
	dat, err := json.Marshal(ms)
	if err != nil {
		panic(err)
	}

	if funcFile, err := os.Create("./" + fileName + ".func"); err != nil {
		panic(err)
	} else {
		fmt.Fprintln(funcFile, "-read")
		fmt.Fprintln(funcFile, string(read))
		fmt.Fprintln(funcFile, "-write")
		fmt.Fprintln(funcFile, string(write))
		funcFile.Sync()
		funcFile.Close()
	}

	if jsonFile, err := os.Create("./" + fileName + ".json"); err != nil {
		panic(err)
	} else {
		fmt.Fprint(jsonFile, string(dat))
		jsonFile.Sync()
		jsonFile.Close()
	}

	if binaryFile, err := os.Create("./" + fileName + ".bin"); err != nil {
		panic(err)
	} else {
		fmt.Fprintf(binaryFile, "var "+fileName+" = []byte{")
		for i, v := range dat {
			if i == len(dat)-1 {
				fmt.Fprintf(binaryFile, "%v", v)
			} else {
				fmt.Fprintf(binaryFile, "%v, ", v)
			}
			if i%31 == 30 {
				fmt.Fprintf(binaryFile, "\n")
			}
		}
		fmt.Fprintf(binaryFile, "}")
		binaryFile.Sync()
		binaryFile.Close()
	}

}
