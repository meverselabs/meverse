package pof

import (
	"bytes"
	"encoding/hex"
	"reflect"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/binutil"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/encoding"
)

func init() {
	encoding.Register(Rank{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(Rank)
		if err := enc.Encode(item.Address); err != nil {
			return err
		}
		if err := enc.Encode(item.PublicHash); err != nil {
			return err
		}
		if err := enc.EncodeUint32(item.phase); err != nil {
			return err
		}
		if err := enc.Encode(item.hashSpace); err != nil {
			return err
		}
		return nil
	}, func(dec *encoding.Decoder, rv reflect.Value) error {
		item := &Rank{}
		if err := dec.Decode(&item.Address); err != nil {
			return err
		}
		if err := dec.Decode(&item.PublicHash); err != nil {
			return err
		}
		if v, err := dec.DecodeUint32(); err != nil {
			return err
		} else {
			item.phase = v
		}
		if err := dec.Decode(&item.hashSpace); err != nil {
			return err
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

// Rank represents the rank information of the formulator account
type Rank struct {
	Address    common.Address
	PublicHash common.PublicHash
	phase      uint32
	hashSpace  hash.Hash256
	score      uint64
}

// NewRank returns a Rank
func NewRank(Address common.Address, PublicHash common.PublicHash, phase uint32, hashSpace hash.Hash256) *Rank {
	m := &Rank{
		phase:     phase,
		hashSpace: hashSpace,
	}
	copy(m.Address[:], Address[:])
	copy(m.PublicHash[:], PublicHash[:])
	m.update()
	return m
}

// Clone returns the clonend value of it
func (rank *Rank) Clone() *Rank {
	return NewRank(rank.Address, rank.PublicHash, rank.phase, rank.hashSpace)
}

// Score returns the score of the rank
func (rank *Rank) Score() uint64 {
	return rank.score
}

// Phase returns the phase of the rank
func (rank *Rank) Phase() uint32 {
	return rank.phase
}

// HashSpace returns the hash space of the rank
func (rank *Rank) HashSpace() hash.Hash256 {
	return rank.hashSpace
}

// Less returns a < b
func (rank *Rank) Less(b *Rank) bool {
	return rank.score < b.score || (rank.score == b.score && bytes.Compare(rank.Address[:], b.Address[:]) < 0)
}

// Equal checks that two values is same or not
func (rank *Rank) Equal(b *Rank) bool {
	return rank.score == b.score && bytes.Equal(rank.Address[:], b.Address[:])
}

// IsZero returns a == 0
func (rank *Rank) IsZero() bool {
	var zeroAddr common.Address
	return rank.score == 0 && bytes.Compare(rank.Address[:], zeroAddr[:]) == 0
}

// Set updates rank's properties and update the score
func (rank *Rank) Set(phase uint32, hashSpace hash.Hash256) {
	rank.phase = phase
	rank.hashSpace = hashSpace
	rank.update()
}

// SetPhase set the phase and update the score
func (rank *Rank) SetPhase(phase uint32) {
	rank.phase = phase
	rank.update()
}

// SetHashSpace set the hash space and update the score
func (rank *Rank) SetHashSpace(hashSpace hash.Hash256) {
	rank.hashSpace = hashSpace
	rank.update()
}

func (rank *Rank) update() {
	rank.score = uint64(rank.phase)<<32 + uint64(binutil.LittleEndian.Uint32(rank.hashSpace[:4]))
}

// Key returns unique key of the rank
func (rank *Rank) Key() string {
	bs := binutil.LittleEndian.Uint64ToBytes(rank.score)
	return string(rank.Address[:]) + "," + string(bs)
}

// String returns the string of the rank using the byte array of rank value
func (rank *Rank) String() string {
	bs := binutil.LittleEndian.Uint64ToBytes(rank.score)
	return rank.Address.String() + "," + hex.EncodeToString(bs)
}
