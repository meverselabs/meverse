package pof

import (
	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
)

// RoundVote is a message for a round vote
type RoundVote struct {
	ChainCoord           *common.Coordinate
	LastHash             hash.Hash256
	VoteTargetHeight     uint32
	TimeoutCount         uint32
	Formulator           common.Address
	FormulatorPublicHash common.PublicHash
	RemainBlocks         uint32
	Timestamp            uint64
	IsReply              bool
}

// RoundVoteMessage is a message for a round vote
type RoundVoteMessage struct {
	RoundVote *RoundVote
	Signautre common.Signature
}

// RoundVoteAck is a message for a round vote ack
type RoundVoteAck struct {
	VoteTargetHeight     uint32
	TimeoutCount         uint32
	Formulator           common.Address
	FormulatorPublicHash common.PublicHash
	RemainBlocks         uint32
	PublicHash           common.PublicHash
	Timestamp            uint64
	IsReply              bool
}

// RoundVoteAckMessage is a message for a round vote
type RoundVoteAckMessage struct {
	RoundVoteAck *RoundVoteAck
	Signautre    common.Signature
}

// BlockVote is message for a block vote
type BlockVote struct {
	VoteTargetHeight   uint32
	Header             *types.Header
	GeneratorSignature common.Signature
	ObserverSignature  common.Signature
	IsReply            bool
}

// BlockVoteMessage is a message for a round vote
type BlockVoteMessage struct {
	BlockVote *BlockVote
	Signautre common.Signature
}

// RoundSetup is a message for a round setup
type RoundSetup struct {
	RoundVoteAcks []*RoundVoteAck
}

// BlockReqMessage is a message for a block request
type BlockReqMessage struct {
	PrevHash             hash.Hash256
	TargetHeight         uint32
	TimeoutCount         uint32
	Formulator           common.Address
	FormulatorPublicHash common.PublicHash
}

// BlockGenMessage is a message for a block generation
type BlockGenMessage struct {
	Block              *types.Block
	GeneratorSignature common.Signature
	IsReply            bool
}

// BlockObSignMessage is a message for a block observer signatures
type BlockObSignMessage struct {
	TargetHeight       uint32
	BlockSign          *types.BlockSign
	ObserverSignatures []common.Signature
}
