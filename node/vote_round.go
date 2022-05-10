package node

import (
	"bytes"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/core/types"
)

type RoundState int

// consts
const (
	EmptyState RoundState = iota
	RoundVoteState
	RoundVoteAckState
	BlockWaitState
	BlockVoteState
)

var RoundStates = [...]string{
	"EmptyState       ",
	"RoundVoteState   ",
	"RoundVoteAckState",
	"BlockWaitState   ",
	"BlockVoteState   ",
}

func (r RoundState) String() string { return RoundStates[(r - 1)] }

// VoteRound is data for the voting round
type VoteRound struct {
	RoundState                 RoundState
	TargetHeight               uint32
	RoundVoteMessageMap        map[common.PublicKey]*RoundVoteMessage
	RoundVoteAckMessageMap     map[common.PublicKey]*RoundVoteAckMessage
	MinRoundVoteAck            *RoundVoteAckMessage
	BlockRoundMap              map[uint32]*BlockRound
	RoundVoteWaitMap           map[common.PublicKey]*RoundVoteMessage
	RoundVoteAckMessageWaitMap map[common.PublicKey]*RoundVoteAckMessage
	VoteFailCount              int
}

// NewVoteRound returns a VoteRound
func NewVoteRound(TargetHeight uint32, MaxBlocksPerGenerator uint32) *VoteRound {
	vr := &VoteRound{
		RoundState:                 RoundVoteState,
		TargetHeight:               TargetHeight,
		RoundVoteMessageMap:        map[common.PublicKey]*RoundVoteMessage{},
		RoundVoteAckMessageMap:     map[common.PublicKey]*RoundVoteAckMessage{},
		RoundVoteWaitMap:           map[common.PublicKey]*RoundVoteMessage{},
		RoundVoteAckMessageWaitMap: map[common.PublicKey]*RoundVoteAckMessage{},
		BlockRoundMap:              map[uint32]*BlockRound{},
	}
	for i := TargetHeight; i < TargetHeight+MaxBlocksPerGenerator; i++ {
		vr.BlockRoundMap[i] = NewBlockRound()
	}
	return vr
}

// BlockRound is data for the block round
type BlockRound struct {
	BlockVoteMap            map[common.PublicKey]*BlockVoteMessage
	BlockGenMessage         *BlockGenMessage
	Context                 *types.Context
	BlockVoteMessageWaitMap map[common.PublicKey]*BlockVoteMessage
	BlockGenMessageWait     *BlockGenMessage
	LastBlockGenRequestTime uint64
}

// NewBlockRound returns a VoteRound
func NewBlockRound() *BlockRound {
	vr := &BlockRound{
		BlockVoteMap:            map[common.PublicKey]*BlockVoteMessage{},
		BlockVoteMessageWaitMap: map[common.PublicKey]*BlockVoteMessage{},
	}
	return vr
}

type voteSortItem struct {
	PublicKey common.PublicKey
	Priority  uint64
}

type voteSorter []*voteSortItem

func (s voteSorter) Len() int {
	return len(s)
}

func (s voteSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s voteSorter) Less(i, j int) bool {
	a := s[i]
	b := s[j]
	if a.Priority == b.Priority {
		return bytes.Compare(a.PublicKey[:], b.PublicKey[:]) < 0
	} else {
		return a.Priority < b.Priority
	}
}
