package types

import (
	"io"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	etypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/pkg/errors"
)

// Receipts implements DerivableList for receipts.
type Receipts []*etypes.Receipt

// Len returns the number of receipts in this list.
func (rs *Receipts) Len() int { return len(*rs) }

// chain.(*Store).StoreBlock
func (rs *Receipts) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	//length
	if sum, err := sw.Uint16(w, uint16(rs.Len())); err != nil {
		return sum, err
	}
	//receipt
	for _, v := range *rs {
		r := newReceiptForStorage(v)
		if sum, err := sw.WriterTo(w, r); err != nil {
			return sum, err
		}
	}
	return sw.Sum(), nil
}

// chain.(*Store).Block
func (rs *Receipts) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()

	if Len, sum, err := sr.GetUint16(r); err != nil {
		return sum, err
	} else {
		for i := uint16(0); i < Len; i++ {
			var v receiptForStorage
			if sum, err := sr.ReaderFrom(r, &v); err != nil {
				return sum, err
			}
			*rs = append(*rs, v.newReceipt())
		}
	}

	return sr.Sum(), nil
}

// Receipt : Status uint64, CumulativeGasUsed uint64, Logs []*Log 만 저장
// 그외 데이터는 block과 transaction 으로부터 구할 수 있다
type receiptForStorage struct {
	Status            uint64
	CumulativeGasUsed uint64
	Logs              []*etypes.Log
}

func newReceiptForStorage(receipt *etypes.Receipt) *receiptForStorage {
	r := receiptForStorage{
		Status:            receipt.Status,
		CumulativeGasUsed: receipt.CumulativeGasUsed,
		Logs:              make([]*etypes.Log, len(receipt.Logs)),
	}
	copy(r.Logs, receipt.Logs)
	return &r
}

func (r *receiptForStorage) newReceipt() *etypes.Receipt {
	receipt := etypes.Receipt{
		Status:            r.Status,
		CumulativeGasUsed: r.CumulativeGasUsed,
		Logs:              make([]*etypes.Log, len(r.Logs)),
	}
	copy(receipt.Logs, r.Logs)
	return &receipt
}

func (r *receiptForStorage) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Uint64(w, r.Status); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint64(w, r.CumulativeGasUsed); err != nil {
		return sum, err
	}
	// Logs
	if sum, err := sw.Uint16(w, uint16(len(r.Logs))); err != nil {
		return sum, err
	}
	for _, v := range r.Logs {
		l := newLogForStroage(v)
		if sum, err := sw.WriterTo(w, l); err != nil {
			return sum, err
		}
	}
	return sw.Sum(), nil
}

// Receipt : Type uint8, Status uint64, CumulativeGasUsed uint64, Logs []*Log, ContractAddress common.Address
// 그외 데이터는 block과 transaction 으로부터 구할 수 있다
func (t *receiptForStorage) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Uint64(r, &t.Status); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint64(r, &t.CumulativeGasUsed); err != nil {
		return sum, err
	}
	// Logs
	if Len, sum, err := sr.GetUint16(r); err != nil {
		return sum, err
	} else {
		t.Logs = make([]*etypes.Log, 0, Len)
		for i := uint16(0); i < Len; i++ {
			v := new(logForStorage)
			if sum, err := sr.ReaderFrom(r, v); err != nil {
				return sum, err
			}
			t.Logs = append(t.Logs, v.newLog())
		}
	}
	return sr.Sum(), nil
}

// Log : Address common.Address, Topics []common.Hash, Data []byte, BlockNumber, Removed bool
// 그외 데이터는 block과 transaction 으로부터 구할 수 있다.
type logForStorage struct {
	Address common.Address
	Topics  []common.Hash
	Data    []byte
	Removed bool
}

func newLogForStroage(log *etypes.Log) *logForStorage {
	l := logForStorage{
		Address: log.Address,
		Topics:  make([]common.Hash, len(log.Topics)),
		Data:    make([]byte, len(log.Data)),
		Removed: log.Removed,
	}
	copy(l.Topics, log.Topics)
	copy(l.Data, log.Data)

	return &l
}

func (l *logForStorage) newLog() *etypes.Log {
	log := etypes.Log{
		Address: l.Address,
		Topics:  make([]common.Hash, len(l.Topics)),
		Data:    make([]byte, len(l.Data)),
		Removed: l.Removed,
	}
	copy(log.Topics, l.Topics)
	copy(log.Data, l.Data)

	return &log
}

func (l *logForStorage) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()

	if sum, err := sw.Address(w, l.Address); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint16(w, uint16(len(l.Topics))); err != nil {
		return sum, err
	}
	for _, v := range l.Topics {
		if sum, err := sw.Hash256(w, v); err != nil {
			return sum, err
		}
	}
	if sum, err := sw.Bytes(w, l.Data); err != nil {
		return sum, err
	}

	if sum, err := sw.Bool(w, l.Removed); err != nil {
		return sum, err
	}

	return sw.Sum(), nil
}

func (l *logForStorage) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()

	if sum, err := sr.Address(r, &l.Address); err != nil {
		return sum, err
	}
	if Len, sum, err := sr.GetUint16(r); err != nil {
		return sum, err
	} else {
		l.Topics = make([]common.Hash, 0, Len)
		for i := uint16(0); i < Len; i++ {
			v := new(common.Hash)
			if sum, err := sr.Hash256(r, v); err != nil {
				return sum, err
			}
			l.Topics = append(l.Topics, *v)
		}
	}
	if sum, err := sr.Bytes(r, &l.Data); err != nil {
		return sum, err
	}
	if sum, err := sr.Bool(r, &l.Removed); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}

// Receipt와 Log 데이터 중 일부를 block과 transaction 로부터 유추
func (rs *Receipts) DeriveReceiptFields(bHash common.Hash, number uint64, idx uint16, tx *etypes.Transaction, signer etypes.Signer) error {

	logIndex := uint(0)
	if idx >= uint16(rs.Len()) {
		return errors.New("transaction and receipt count mismatch")
	}

	r := (*rs)[idx]
	r.Type = tx.Type()
	r.TxHash = tx.Hash()

	// block location fields
	r.BlockHash = bHash
	r.BlockNumber = new(big.Int).SetUint64(number)
	r.TransactionIndex = uint(idx)

	// The contract address can be derived from the transaction itself
	if tx.To() == nil {
		// Deriving the signer is expensive, only do if it's actually needed
		from, _ := etypes.Sender(signer, tx)
		r.ContractAddress = crypto.CreateAddress(from, tx.Nonce())
	}

	// The used gas can be calculated based on previous r
	if idx == 0 {
		r.GasUsed = r.CumulativeGasUsed
	} else {
		r.GasUsed = r.CumulativeGasUsed - (*rs)[idx-1].CumulativeGasUsed
	}
	// The derived log fields can simply be set from the block and transaction
	for j := 0; j < len(r.Logs); j++ {
		r.Logs[j].BlockNumber = number
		r.Logs[j].BlockHash = bHash
		r.Logs[j].TxHash = r.TxHash
		r.Logs[j].TxIndex = uint(idx)
		r.Logs[j].Index = logIndex
		logIndex++
	}

	return nil
}
