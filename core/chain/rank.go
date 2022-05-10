package chain

import (
	"bytes"
	"encoding/hex"
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
)

// Rank represents the rank information of the formulator account
type Rank struct {
	Address   common.Address
	phase     uint32
	hashSpace hash.Hash256
	score     uint64
}

// NewRank returns a Rank
func NewRank(Address common.Address, phase uint32, hashSpace hash.Hash256) *Rank {
	m := &Rank{
		phase:     phase,
		hashSpace: hashSpace,
	}
	copy(m.Address[:], Address[:])
	m.update()
	return m
}

// Clone returns the clonend value of it
func (rank *Rank) Clone() *Rank {
	return NewRank(rank.Address, rank.phase, rank.hashSpace)
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
	var emptyAddr common.Address
	return rank.score == 0 && rank.Address == emptyAddr
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
	rank.score = uint64(rank.phase)<<32 + uint64(bin.Uint32(rank.hashSpace[:4]))
}

// Key returns unique key of the rank
func (rank *Rank) Key() string {
	bs := bin.Uint64Bytes(rank.score)
	return string(rank.Address[:]) + "," + string(bs)
}

// String returns the string of the rank using the byte array of rank value
func (rank *Rank) String() string {
	bs := bin.Uint64Bytes(rank.score)
	return rank.Address.String() + "," + hex.EncodeToString(bs)
}

func (s *Rank) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Address(w, s.Address); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.phase); err != nil {
		return sum, err
	}
	if sum, err := sw.Hash256(w, s.hashSpace); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint64(w, s.score); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *Rank) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Address(r, &s.Address); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.phase); err != nil {
		return sum, err
	}
	if sum, err := sr.Hash256(r, &s.hashSpace); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint64(r, &s.score); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
