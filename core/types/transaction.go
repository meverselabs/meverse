package types

import (
	"fmt"
	"io"
	"math/big"
	"strconv"

	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/extern/txparser"
)

type Transaction struct {
	//Input data
	ChainID     *big.Int
	Timestamp   uint64
	Seq         uint64
	To          common.Address
	Method      string
	Args        []byte
	GasPrice    *big.Int
	UseSeq      bool
	IsEtherType bool

	//After exec tx data
	From common.Address
}

var legacyCheckHeight uint32

func SetLegacyCheckHeight(l uint32) {
	legacyCheckHeight = l
}

func NewTransaction(ctx *Context, _timeStamp string, _to common.Address, _method string, _args []byte) (*Transaction, error) {
	timeStamp, err := strconv.ParseUint(_timeStamp, 10, 64)
	if err != nil {
		return nil, err
	}

	tx := &Transaction{
		ChainID:   ctx.ChainID(),
		Timestamp: timeStamp,
		To:        _to,
		Method:    _method,
		Args:      _args,
	}

	return tx, nil
}

func (s *Transaction) withOutFrom() *Transaction {
	return &Transaction{
		ChainID:     big.NewInt(0).SetBytes(s.ChainID.Bytes()),
		Timestamp:   s.Timestamp,
		Seq:         s.Seq,
		From:        common.Address{},
		To:          s.To,
		Method:      s.Method,
		Args:        s.Args,
		IsEtherType: s.IsEtherType,
		GasPrice:    s.GasPrice,
	}
}

func (s *Transaction) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.BigInt(w, s.ChainID); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint64(w, s.Timestamp); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint64(w, s.Seq); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.From); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.To); err != nil {
		return sum, err
	}
	if sum, err := sw.String(w, s.Method); err != nil {
		return sum, err
	}
	if sum, err := sw.Bytes(w, s.Args); err != nil {
		return sum, err
	}
	if sum, err := sw.BigInt(w, s.GasPrice); err != nil {
		return sum, err
	}
	if sum, err := sw.Bool(w, s.UseSeq); err != nil {
		return sum, err
	}
	if sum, err := sw.Bool(w, s.IsEtherType); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *Transaction) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.BigInt(r, &s.ChainID); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint64(r, &s.Timestamp); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint64(r, &s.Seq); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.From); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.To); err != nil {
		return sum, err
	}
	if sum, err := sr.String(r, &s.Method); err != nil {
		return sum, err
	}
	if sum, err := sr.Bytes(r, &s.Args); err != nil {
		return sum, err
	}
	if sum, err := sr.BigInt(r, &s.GasPrice); err != nil {
		return sum, err
	}
	if sum, err := sr.Bool(r, &s.UseSeq); err != nil {
		return sum, err
	}
	if sum, err := sr.Bool(r, &s.IsEtherType); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}

func (s *Transaction) Hash(height uint32) (h hash.Hash256) {
	if s.IsEtherType {
		etx, _, err := txparser.EthTxFromRLP(s.Args)
		if err != nil {
			fmt.Println("Transaction hash read Value error")
			return
		}
		if height < legacyCheckHeight {
			signer := etypes.NewLondonSigner(etx.ChainId())
			h = signer.Hash(etx)
		} else {
			h = etx.Hash()
		}
	} else {
		var err error
		h, _, err = bin.WriterToHash(s.withOutFrom())
		if err != nil {
			fmt.Println("Transaction hash error", err)
		}
	}
	return
}

func (s *Transaction) HashSig() (h hash.Hash256) {
	if s.IsEtherType {
		etx, _, err := txparser.EthTxFromRLP(s.Args)
		if err != nil {
			fmt.Println("Transaction hash read Value error")
			return
		}
		h = etx.Hash()
	} else {
		var err error
		h, _, err = bin.WriterToHash(s.withOutFrom())
		if err != nil {
			fmt.Println("Transaction hash error", err)
		}
	}
	return
}

func (s *Transaction) Message() (h hash.Hash256) {
	if s.IsEtherType {
		etx, _, err := txparser.EthTxFromRLP(s.Args)
		if err != nil {
			fmt.Println("Transaction hash read Value error")
			return
		}
		signer := etypes.NewLondonSigner(etx.ChainId())
		h = signer.Hash(etx)
	} else {
		var err error
		h, _, err = bin.WriterToHash(s.withOutFrom())
		if err != nil {
			fmt.Println("Transaction hash error", err)
		}
	}
	return
}

func TxArg(ctx *Context, tx *Transaction) (data []interface{}, isSendValue bool, err error) {
	if tx.IsEtherType {
		var etx *etypes.Transaction
		etx, _, err = txparser.EthTxFromRLP(tx.Args)
		if err == nil {
			eData := etx.Data()
			if etx.Value().Cmp(amount.ZeroCoin.Int) > 0 && tx.To != *ctx.MainToken() {
				if len(eData) > 0 {
					err = errors.New("not support value transfer call")
				} else {
					isSendValue = true
					data = []interface{}{&amount.Amount{Int: etx.Value()}}
				}
			} else if len(eData) > 0 {
				data, err = txparser.Inputs(eData)
			} else {
				err = errors.New("invalid call")
			}
		}
	} else {
		data, err = bin.TypeReadAll(tx.Args, -1)
	}
	return
}
