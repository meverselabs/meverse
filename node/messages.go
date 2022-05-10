package node

import (
	"io"
	"math/big"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
	"github.com/meverselabs/meverse/p2p"
)

var (
	RoundVoteMessageType       = p2p.RegisterSerializableType(&RoundVoteMessage{})
	RoundVoteAckMessageType    = p2p.RegisterSerializableType(&RoundVoteAckMessage{})
	BlockReqMessageType        = p2p.RegisterSerializableType(&BlockReqMessage{})
	BlockGenMessageType        = p2p.RegisterSerializableType(&BlockGenMessage{})
	BlockVoteMessageType       = p2p.RegisterSerializableType(&BlockVoteMessage{})
	BlockObSignMessageType     = p2p.RegisterSerializableType(&BlockObSignMessage{})
	BlockGenRequestMessageType = p2p.RegisterSerializableType(&BlockGenRequestMessage{})
)

// RoundVoteMessage is a message for a round vote
type RoundVoteMessage struct {
	ChainID      *big.Int
	LastHash     hash.Hash256
	TargetHeight uint32
	TimeoutCount uint32
	Generator    common.Address
	PublicKey    common.PublicKey
	Timestamp    uint64
	IsReply      bool
}

func (s *RoundVoteMessage) TypeID() uint32 {
	return RoundVoteMessageType
}

func (s *RoundVoteMessage) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Bytes(w, s.ChainID.Bytes()); err != nil {
		return sum, err
	}
	if sum, err := sw.Hash256(w, s.LastHash); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.TargetHeight); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.TimeoutCount); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Generator); err != nil {
		return sum, err
	}
	if sum, err := sw.PublicKey(w, s.PublicKey); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint64(w, s.Timestamp); err != nil {
		return sum, err
	}
	if sum, err := sw.Bool(w, s.IsReply); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *RoundVoteMessage) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	cibs := []byte{}
	if sum, err := sr.Bytes(r, &cibs); err != nil {
		return sum, err
	}
	s.ChainID = big.NewInt(0).SetBytes(cibs)
	if sum, err := sr.Hash256(r, &s.LastHash); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.TargetHeight); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.TimeoutCount); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Generator); err != nil {
		return sum, err
	}
	if sum, err := sr.PublicKey(r, &s.PublicKey); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint64(r, &s.Timestamp); err != nil {
		return sum, err
	}
	if sum, err := sr.Bool(r, &s.IsReply); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}

// RoundVoteAckMessage is a message for a round vote ack
type RoundVoteAckMessage struct {
	ChainID      *big.Int
	LastHash     hash.Hash256
	TargetHeight uint32
	TimeoutCount uint32
	Generator    common.Address
	PublicKey    common.PublicKey
	Timestamp    uint64
	IsReply      bool
}

func (s *RoundVoteAckMessage) TypeID() uint32 {
	return RoundVoteAckMessageType
}

func (s *RoundVoteAckMessage) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Bytes(w, s.ChainID.Bytes()); err != nil {
		return sum, err
	}
	if sum, err := sw.Hash256(w, s.LastHash); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.TargetHeight); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.TimeoutCount); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Generator); err != nil {
		return sum, err
	}
	if sum, err := sw.PublicKey(w, s.PublicKey); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint64(w, s.Timestamp); err != nil {
		return sum, err
	}
	if sum, err := sw.Bool(w, s.IsReply); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *RoundVoteAckMessage) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	cibs := []byte{}
	if sum, err := sr.Bytes(r, &cibs); err != nil {
		return sum, err
	}
	s.ChainID = big.NewInt(0).SetBytes(cibs)
	if sum, err := sr.Hash256(r, &s.LastHash); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.TargetHeight); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.TimeoutCount); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Generator); err != nil {
		return sum, err
	}
	if sum, err := sr.PublicKey(r, &s.PublicKey); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint64(r, &s.Timestamp); err != nil {
		return sum, err
	}
	if sum, err := sr.Bool(r, &s.IsReply); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}

// BlockReqMessage is a message for a block request
type BlockReqMessage struct {
	PrevHash     hash.Hash256
	TargetHeight uint32
	TimeoutCount uint32
	Generator    common.Address
}

func (s *BlockReqMessage) TypeID() uint32 {
	return BlockReqMessageType
}

func (s *BlockReqMessage) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Hash256(w, s.PrevHash); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.TargetHeight); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.TimeoutCount); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Generator); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *BlockReqMessage) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Hash256(r, &s.PrevHash); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.TargetHeight); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.TimeoutCount); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Generator); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}

// BlockGenMessage is a message for a block generation
type BlockGenMessage struct {
	Block              *types.Block
	GeneratorSignature common.Signature
	IsReply            bool
}

func (s *BlockGenMessage) TypeID() uint32 {
	return BlockGenMessageType
}

func (s *BlockGenMessage) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.WriterTo(w, s.Block); err != nil {
		return sum, err
	}
	if sum, err := sw.Signature(w, s.GeneratorSignature); err != nil {
		return sum, err
	}
	if sum, err := sw.Bool(w, s.IsReply); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *BlockGenMessage) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	s.Block = &types.Block{}
	if sum, err := sr.ReaderFrom(r, s.Block); err != nil {
		return sum, err
	}
	if sum, err := sr.Signature(r, &s.GeneratorSignature); err != nil {
		return sum, err
	}
	if sum, err := sr.Bool(r, &s.IsReply); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}

// BlockVoteMessage is message for a block vote
type BlockVoteMessage struct {
	TargetHeight       uint32
	Header             *types.Header
	GeneratorSignature common.Signature
	ObserverSignature  common.Signature
	IsReply            bool
}

func (s *BlockVoteMessage) TypeID() uint32 {
	return BlockVoteMessageType
}

func (s *BlockVoteMessage) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Uint32(w, s.TargetHeight); err != nil {
		return sum, err
	}
	if sum, err := sw.WriterTo(w, s.Header); err != nil {
		return sum, err
	}
	if sum, err := sw.Signature(w, s.GeneratorSignature); err != nil {
		return sum, err
	}
	if sum, err := sw.Signature(w, s.ObserverSignature); err != nil {
		return sum, err
	}
	if sum, err := sw.Bool(w, s.IsReply); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *BlockVoteMessage) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Uint32(r, &s.TargetHeight); err != nil {
		return sum, err
	}
	s.Header = &types.Header{}
	if sum, err := sr.ReaderFrom(r, s.Header); err != nil {
		return sum, err
	}
	if sum, err := sr.Signature(r, &s.GeneratorSignature); err != nil {
		return sum, err
	}
	if sum, err := sr.Signature(r, &s.ObserverSignature); err != nil {
		return sum, err
	}
	if sum, err := sr.Bool(r, &s.IsReply); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}

// BlockObSignMessage is a message for a block observer signatures
type BlockObSignMessage struct {
	TargetHeight       uint32
	BlockSign          *types.BlockSign
	ObserverSignatures []common.Signature
}

func (s *BlockObSignMessage) TypeID() uint32 {
	return BlockObSignMessageType
}

func (s *BlockObSignMessage) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Uint32(w, s.TargetHeight); err != nil {
		return sum, err
	}
	if sum, err := sw.WriterTo(w, s.BlockSign); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint16(w, uint16(len(s.ObserverSignatures))); err != nil {
		return sum, err
	}
	for _, v := range s.ObserverSignatures {
		if sum, err := sw.Signature(w, v); err != nil {
			return sum, err
		}
	}
	return sw.Sum(), nil
}

func (s *BlockObSignMessage) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Uint32(r, &s.TargetHeight); err != nil {
		return sum, err
	}
	s.BlockSign = &types.BlockSign{}
	if sum, err := sr.ReaderFrom(r, s.BlockSign); err != nil {
		return sum, err
	}
	if Len, sum, err := sr.GetUint16(r); err != nil {
		return sum, err
	} else {
		s.ObserverSignatures = make([]common.Signature, 0, Len)
		for i := uint16(0); i < Len; i++ {
			var v common.Signature
			if sum, err := sr.Signature(r, &v); err != nil {
				return sum, err
			}
			s.ObserverSignatures = append(s.ObserverSignatures, v)
		}
	}
	return sr.Sum(), nil
}

// BlockGenRequestMessage is a message to request block gen
type BlockGenRequestMessage struct {
	ChainID      *big.Int
	LastHash     hash.Hash256
	TargetHeight uint32
	TimeoutCount uint32
	Generator    common.Address
	Timestamp    uint64
}

func (s *BlockGenRequestMessage) TypeID() uint32 {
	return BlockGenRequestMessageType
}

func (s *BlockGenRequestMessage) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Bytes(w, s.ChainID.Bytes()); err != nil {
		return sum, err
	}
	if sum, err := sw.Hash256(w, s.LastHash); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.TargetHeight); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.TimeoutCount); err != nil {
		return sum, err
	}
	if sum, err := sw.Address(w, s.Generator); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint64(w, s.Timestamp); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *BlockGenRequestMessage) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	cibs := []byte{}
	if sum, err := sr.Bytes(r, &cibs); err != nil {
		return sum, err
	}
	s.ChainID = big.NewInt(0).SetBytes(cibs)
	if sum, err := sr.Hash256(r, &s.LastHash); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.TargetHeight); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.TimeoutCount); err != nil {
		return sum, err
	}
	if sum, err := sr.Address(r, &s.Generator); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint64(r, &s.Timestamp); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}
