package consensus

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/util"
)

// Rank represents the rank information of the formulation account
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

// WriteTo is a serialization function
func (rank *Rank) WriteTo(w io.Writer) (int64, error) {
	var wrote int64
	if n, err := rank.Address.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := rank.PublicHash.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := util.WriteUint32(w, rank.phase); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	if n, err := rank.hashSpace.WriteTo(w); err != nil {
		return wrote, err
	} else {
		wrote += n
	}
	return wrote, nil
}

// ReadFrom is a deserialization function
func (rank *Rank) ReadFrom(r io.Reader) (int64, error) {
	var read int64
	if n, err := rank.Address.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if n, err := rank.PublicHash.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	if v, n, err := util.ReadUint32(r); err != nil {
		return read, err
	} else {
		read += n
		rank.phase = v
	}
	if n, err := rank.hashSpace.ReadFrom(r); err != nil {
		return read, err
	} else {
		read += n
	}
	rank.update()
	return read, nil
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
	rank.score = uint64(rank.phase)<<32 + uint64(binary.LittleEndian.Uint32(rank.hashSpace[:4]))
}

// Key returns unique key of the rank
func (rank *Rank) Key() string {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, rank.score)
	return string(rank.Address[:]) + "," + string(bs)
}

// String returns the string of the rank using the byte array of rank value
func (rank *Rank) String() string {
	bs := make([]byte, 8)
	binary.LittleEndian.PutUint64(bs, rank.score)
	return string(rank.PublicHash[:]) + "," + string(bs)
}
