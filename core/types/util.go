package types

import (
	"bytes"
	"encoding/hex"
	"sort"
	"time"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/amount"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/pkg/errors"
)

// UnmarshalID returns the block height, the transaction index in the block, the output index in the transaction
func UnmarshalID(id uint64) (uint32, uint16, uint16) {
	return uint32(id >> 32), uint16(id >> 16), uint16(id)
}

// MarshalID returns the packed id
func MarshalID(height uint32, index uint16, n uint16) uint64 {
	return uint64(height)<<32 | uint64(index)<<16 | uint64(n)
}

// TransactionIDBytes returns the id bytes of the transaction
func TransactionIDBytes(Height uint32, Index uint16) []byte {
	bs := make([]byte, 6)
	bin.PutUint32(bs, Height)
	bin.PutUint16(bs[4:], Index)
	return bs
}

// TransactionID returns the id of the transaction
func TransactionID(Height uint32, Index uint16) string {
	bs := TransactionIDBytes(Height, Index)
	return hex.EncodeToString(bs)
}

// ParseTransactionID returns the id of the transaction
func ParseTransactionIDBytes(bs []byte) (uint32, uint16, error) {
	if len(bs) != 6 {
		return 0, 0, errors.WithStack(ErrInvalidTransactionIDFormat)
	}
	Height := bin.Uint32(bs)
	Index := bin.Uint16(bs[4:])
	return Height, Index, nil
}

// ParseTransactionID returns the id of the transaction
func ParseTransactionID(TXID string) (uint32, uint16, error) {
	if len(TXID) != 12 {
		return 0, 0, errors.WithStack(ErrInvalidTransactionIDFormat)
	}
	bs, err := hex.DecodeString(TXID)
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}
	return ParseTransactionIDBytes(bs)
}

// ToTimeSlot returns the timeslot of the timestamp
func ToTimeSlot(timestamp uint64) uint32 {
	return uint32(timestamp / uint64(60*time.Second))
}

func EachAllAddressBool(mp map[common.Address]bool, fn func(key common.Address, value bool) error) error {
	keys := []common.Address{}
	for k := range mp {
		keys = append(keys, k)
	}
	sort.Sort(addressListSorter(keys))
	for _, k := range keys {
		v := mp[k]
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

func EachAllAddressUint32(mp map[common.Address]uint32, fn func(key common.Address, value uint32) error) error {
	keys := []common.Address{}
	for k := range mp {
		keys = append(keys, k)
	}
	sort.Sort(addressListSorter(keys))
	for _, k := range keys {
		v := mp[k]
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

func EachAllAddressUint64(mp map[common.Address]uint64, fn func(key common.Address, value uint64) error) error {
	keys := []common.Address{}
	for k := range mp {
		keys = append(keys, k)
	}
	sort.Sort(addressListSorter(keys))
	for _, k := range keys {
		v := mp[k]
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

func EachAllAddressAmount(mp map[common.Address]*amount.Amount, fn func(key common.Address, value *amount.Amount) error) error {
	keys := []common.Address{}
	for k := range mp {
		keys = append(keys, k)
	}
	sort.Sort(addressListSorter(keys))
	for _, k := range keys {
		v := mp[k]
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

func EachAllAddressBytes(mp map[common.Address][]byte, fn func(key common.Address, value []byte) error) error {
	keys := []common.Address{}
	for k := range mp {
		keys = append(keys, k)
	}
	sort.Sort(addressListSorter(keys))
	for _, k := range keys {
		v := mp[k]
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

func EachAllStringBytes(mp map[string][]byte, fn func(key string, value []byte) error) error {
	keys := []string{}
	for k := range mp {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := mp[k]
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

func EachAllStringBool(mp map[string]bool, fn func(key string, value bool) error) error {
	keys := []string{}
	for k := range mp {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := mp[k]
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

func EachAllTimeSlotMap(mp map[uint32]map[string]bool, fn func(key uint32, value map[string]bool) error) error {
	keys := []uint32{}
	for k := range mp {
		keys = append(keys, k)
	}
	sort.Sort(uint32ListSorter(keys))
	for _, k := range keys {
		v := mp[k]
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

func EachAllAddressContractDefine(mp map[common.Address]*ContractDefine, fn func(key common.Address, cd *ContractDefine) error) error {
	keys := []common.Address{}
	for k := range mp {
		keys = append(keys, k)
	}
	sort.Sort(addressListSorter(keys))
	for _, k := range keys {
		v := mp[k]
		if err := fn(k, v); err != nil {
			return err
		}
	}
	return nil
}

type uint32ListSorter []uint32

func (p uint32ListSorter) Len() int           { return len(p) }
func (p uint32ListSorter) Less(i, j int) bool { return p[i] < p[j] }
func (p uint32ListSorter) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type addressListSorter []common.Address

func (p addressListSorter) Len() int           { return len(p) }
func (p addressListSorter) Less(i, j int) bool { return bytes.Compare(p[i][:], p[j][:]) < 0 }
func (p addressListSorter) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func MarshalAddressAmountMap(mp map[common.Address]*amount.Amount) ([]byte, error) {
	var buffer bytes.Buffer
	sw := bin.NewSumWriter()
	if _, err := sw.Uint32(&buffer, uint32(len(mp))); err != nil {
		return nil, err
	}

	if err := EachAllAddressAmount(mp, func(k common.Address, v *amount.Amount) error {
		if _, err := sw.Address(&buffer, k); err != nil {
			return err
		}
		if _, err := sw.Amount(&buffer, v); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func UnmarshalAddressAmountMap(bs []byte, mp map[common.Address]*amount.Amount) error {
	r := bytes.NewReader(bs)
	sr := bin.NewSumReader()
	if Len, _, err := sr.GetUint32(r); err != nil {
		return err
	} else {
		for i := uint32(0); i < Len; i++ {
			var addr common.Address
			if _, err := sr.Address(r, &addr); err != nil {
				return err
			}
			var am *amount.Amount
			if _, err := sr.Amount(r, &am); err != nil {
				return err
			}
			mp[addr] = am
		}
	}
	return nil
}

func MarshalAddressBytesMap(mp map[common.Address][]byte) ([]byte, error) {
	var buffer bytes.Buffer
	sw := bin.NewSumWriter()
	if _, err := sw.Uint32(&buffer, uint32(len(mp))); err != nil {
		return nil, err
	}

	if err := EachAllAddressBytes(mp, func(k common.Address, v []byte) error {
		if _, err := sw.Address(&buffer, k); err != nil {
			return err
		}
		if _, err := sw.Bytes(&buffer, v); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func UnmarshalAddressBytesMap(bs []byte, mp map[common.Address][]byte) error {
	r := bytes.NewReader(bs)
	sr := bin.NewSumReader()
	if Len, _, err := sr.GetUint32(r); err != nil {
		return err
	} else {
		for i := uint32(0); i < Len; i++ {
			var addr common.Address
			if _, err := sr.Address(r, &addr); err != nil {
				return err
			}
			var data []byte
			if _, err := sr.Bytes(r, &data); err != nil {
				return err
			}
			mp[addr] = data
		}
	}
	return nil
}
