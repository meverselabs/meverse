package pof

import (
	"bytes"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

// Candidates returns a candidates
func (cs *Consensus) Candidates() []*Rank {
	cs.Lock()
	defer cs.Unlock()

	return cs.rt.Candidates()
}

func (cs *Consensus) updateFormulatorList(ctw *types.ContextWrapper) error {
	var inErr error
	phase := cs.rt.smallestPhase() + 2
	if cs.maxPhaseDiff != nil {
		diff := cs.maxPhaseDiff(ctw.TargetHeight())
		phase = cs.rt.largestPhase() + 1
		if diff != 0 {
			maxPhase := cs.rt.smallestPhase() + diff
			if phase > maxPhase {
				phase = maxPhase
			}
			if cs.rt.largestPhase() > maxPhase {
				list := cs.rt.Candidates()
				for _, r := range list {
					if r.phase > maxPhase {
						cs.rt.removeRank(r.Address)
						if err := cs.rt.addRank(NewRank(r.Address, r.PublicHash, maxPhase, r.hashSpace)); err != nil {
							return err
						}
					}
				}
			}
		}
	}
	ctw.Top().AccountMap.EachAll(func(addr common.Address, a types.Account) bool {
		if acc, is := a.(FormulatorAccount); is && acc.IsFormulator() {
			if a.Address().Height() == ctw.TargetHeight() {
				if acc.IsActivated() {
					if err := cs.rt.addRank(NewRank(addr, acc.GeneratorHash(), phase, hash.DoubleHash(addr[:]))); err != nil {
						inErr = err
						return false
					}
				}
			} else {
				r, has := cs.rt.rankMap[addr]
				if has {
					if !acc.IsActivated() {
						cs.rt.removeRank(addr)
					} else {
						if r.PublicHash != acc.GeneratorHash() {
							cs.rt.removeRank(addr)
							if err := cs.rt.addRank(NewRank(addr, acc.GeneratorHash(), phase, hash.DoubleHash(addr[:]))); err != nil {
								inErr = err
								return false
							}
						}
					}
				} else {
					if acc.IsActivated() {
						if err := cs.rt.addRank(NewRank(addr, acc.GeneratorHash(), phase, hash.DoubleHash(addr[:]))); err != nil {
							inErr = err
							return false
						}
					}
				}
			}
		}
		return true
	})
	if inErr != nil {
		return inErr
	}
	ctw.Top().DeletedAccountMap.EachAll(func(addr common.Address, a types.Account) bool {
		if acc, is := a.(FormulatorAccount); is && acc.IsFormulator() {
			cs.rt.removeRank(acc.Address())
		}
		return true
	})
	if cs.rt.CandidateCount() == 0 {
		return ErrInsufficientCandidateCount
	}
	return nil
}

// DecodeConsensusData decodes header's consensus data
func (cs *Consensus) DecodeConsensusData(ConsensusData []byte) (uint32, error) {
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
	if err := enc.EncodeUint32(cs.maxBlocksPerFormulator); err != nil {
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
