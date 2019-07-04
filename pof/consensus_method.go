package pof

import (
	"bytes"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
	"github.com/fletaio/fleta/process/formulator"
)

// Candidates returns a candidates
func (cs *Consensus) Candidates() []*Rank {
	cs.Lock()
	defer cs.Unlock()

	return cs.rt.Candidates()
}

func (cs *Consensus) updateFormulatorList(ctw *types.ContextWrapper) error {
	var inErr error
	phase := cs.rt.largestPhase() + 1
	ctw.Top().AccountMap.EachAll(func(addr common.Address, a types.Account) bool {
		if a.Address().Coordinate().Height == ctw.TargetHeight() {
			if acc, is := a.(*formulator.FormulatorAccount); is {
				addr := acc.Address()
				if err := cs.rt.addRank(NewRank(addr, acc.KeyHash, phase, hash.DoubleHash(addr[:]))); err != nil {
					inErr = err
					return false
				}
			}
		}
		return true
	})
	if inErr != nil {
		return inErr
	}
	ctw.Top().DeletedAccountMap.EachAll(func(addr common.Address, a types.Account) bool {
		if acc, is := a.(*formulator.FormulatorAccount); is {
			cs.rt.removeRank(acc.Address())
		}
		return true
	})
	if cs.rt.CandidateCount() == 0 {
		return ErrInsufficientCandidateCount
	}
	return nil
}

func (cs *Consensus) decodeConsensusData(ConsensusData []byte) (uint32, error) {
	dec := encoding.NewDecoder(bytes.NewReader(ConsensusData))
	TimeoutCount, err := dec.DecodeUint32()
	if err != nil {
		return 0, err
	}
	return TimeoutCount, nil
}

func (cs *Consensus) encodeConsensusData(TimeoutCount uint32) ([]byte, error) {
	var buffer bytes.Buffer
	enc := encoding.NewEncoder(&buffer)
	if err := enc.EncodeUint32(TimeoutCount); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (cs *Consensus) buildSaveData() ([]byte, error) {
	var buffer bytes.Buffer
	enc := encoding.NewEncoder(&buffer)
	if err := enc.EncodeUint32(cs.MaxBlocksPerFormulator); err != nil {
		return nil, err
	}
	if err := enc.Encode(cs.observerKeyMap); err != nil {
		return nil, err
	}
	if err := enc.EncodeUint32(cs.blocksBySameFormulator); err != nil {
		return nil, err
	}
	if err := enc.Encode(cs.rt); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
