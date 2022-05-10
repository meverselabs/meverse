package txparser

import (
	"encoding/hex"
	"log"
	"math/big"
	"testing"

	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/meverselabs/meverse/common"
)

func decode(rlp string) []byte {
	h, err := hex.DecodeString(rlp)
	if err != nil {
		panic(err)
	}
	return h
}

func TestEthTxFromRLP(t *testing.T) {
	type args struct {
		rlpBytes []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid sig",
			args: args{
				rlpBytes: decode("f8ab3685e8d4a51000830249f094ba1611d325f7c85d089a86be0e3e70da69679e9380b844a9059cbb000000000000000000000000ba17b965f8c7caadf289877d07e973f6aa36b0930000000000000000000000000000000000000000000000056bc75e2d63100000822552a0ee92f7a6b5c8798b104329fa152310218242b690b106dc8a9bde0c84a37c4f619f61d3ec872d571532191d4e75e697c95c0c3343259a926bccd1518b315cfaec"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			etx, sig, err := EthTxFromRLP(tt.args.rlpBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("EthTxFromRLP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			signer := etypes.NewLondonSigner(etx.ChainId())
			TxHash := signer.Hash(etx)

			pubkey, err := common.RecoverPubkey(big.NewInt(0x1297), TxHash, sig)
			if (err != nil) != tt.wantErr {
				t.Errorf("EthTxFromRLP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			log.Println(pubkey)

		})
	}
}
