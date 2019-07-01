package consensus

import (
	"bytes"
	"sort"
	"sync"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/util"
	"github.com/fletaio/fleta/core/account"
	"github.com/fletaio/fleta/core/block"
	"github.com/fletaio/fleta/core/data"
)

// Consensus supports the proof of formulation algorithm
type Consensus struct {
	sync.Mutex
	height                   uint64
	candidates               []*Rank
	rankMap                  map[common.Address]*Rank
	blocksFromSameFormulator uint32
	ObserverKeyMap           map[common.PublicHash]bool
	MaxBlocksPerFormulator   uint32
	FormulationAccountType   account.Type
}

// NewConsensus returns a Consensus
func NewConsensus(ObserverKeyMap map[common.PublicHash]bool, MaxBlocksPerFormulator uint32, FormulationAccountType account.Type) *Consensus {
	cs := &Consensus{
		candidates:             []*Rank{},
		rankMap:                map[common.Address]*Rank{},
		ObserverKeyMap:         ObserverKeyMap,
		MaxBlocksPerFormulator: MaxBlocksPerFormulator,
		FormulationAccountType: FormulationAccountType,
	}
	return cs
}

// CandidateCount returns a count of the rank table
func (cs *Consensus) CandidateCount() int {
	cs.Lock()
	defer cs.Unlock()

	return len(cs.candidates)
}

// BlocksFromSameFormulator returns a number of blocks made from same formulator
func (cs *Consensus) BlocksFromSameFormulator() uint32 {
	cs.Lock()
	defer cs.Unlock()

	return cs.blocksFromSameFormulator
}

// TopRank returns the top rank by Timeoutcount
func (cs *Consensus) TopRank(TimeoutCount int) (*Rank, error) {
	cs.Lock()
	defer cs.Unlock()

	if TimeoutCount >= len(cs.candidates) {
		return nil, ErrInsufficientCandidateCount
	}
	return cs.candidates[TimeoutCount].Clone(), nil
}

// TopRankInMap returns the top rank by the given timeout count in the given map
func (cs *Consensus) TopRankInMap(FormulatorMap map[common.Address]bool) (*Rank, int, error) {
	cs.Lock()
	defer cs.Unlock()

	if len(FormulatorMap) == 0 {
		return nil, 0, ErrInsufficientCandidateCount
	}
	for i, r := range cs.candidates {
		if FormulatorMap[r.Address] {
			return r, i, nil
		}
	}
	return nil, 0, ErrInsufficientCandidateCount
}

// RanksInMap returns the ranks in the map
func (cs *Consensus) RanksInMap(FormulatorMap map[common.Address]bool, Limit int) ([]*Rank, error) {
	cs.Lock()
	defer cs.Unlock()

	if len(FormulatorMap) == 0 {
		return nil, ErrInsufficientCandidateCount
	}
	if Limit < 1 {
		Limit = 1
	}
	list := make([]*Rank, 0, Limit)
	for _, r := range cs.candidates {
		if FormulatorMap[r.Address] {
			list = append(list, r)
			if len(list) >= Limit {
				break
			}
		}
	}
	return list, nil
}

// IsFormulator returns the given information is correct or not
func (cs *Consensus) IsFormulator(Formulator common.Address, Publichash common.PublicHash) bool {
	cs.Lock()
	defer cs.Unlock()

	rank := cs.rankMap[Formulator]
	if rank == nil {
		return false
	}
	if !rank.PublicHash.Equal(Publichash) {
		return false
	}
	return true
}

// ApplyGenesis initialize the consensus using the genesis context data
func (cs *Consensus) ApplyGenesis(ctd *data.ContextData) ([]byte, error) {
	cs.Lock()
	defer cs.Unlock()

	phase := cs.largestPhase() + 1
	for _, a := range ctd.CreatedAccountMap {
		if a.Type() == cs.FormulationAccountType {
			acc := a.(*FormulationAccount)
			addr := acc.Address()
			if err := cs.addRank(NewRank(addr, acc.KeyHash, phase, hash.DoubleHash(addr[:]))); err != nil {
				return nil, err
			}
		}
	}
	for _, acc := range ctd.DeletedAccountMap {
		if acc.Type() == cs.FormulationAccountType {
			cs.removeRank(acc.Address())
		}
	}
	SaveData, err := cs.buildSaveData()
	if err != nil {
		return nil, err
	}
	return SaveData, nil
}

// ProcessContext processes the consensus using the block and its context data
func (cs *Consensus) ProcessContext(ctd *data.ContextData, HeaderHash hash.Hash256, bh *block.Header) ([]byte, error) {
	cs.Lock()
	defer cs.Unlock()

	if bh.TimeoutCount > 0 {
		if err := cs.forwardCandidates(int(bh.TimeoutCount)); err != nil {
			return nil, err
		}
		cs.blocksFromSameFormulator = 0
	}
	cs.blocksFromSameFormulator++
	if cs.blocksFromSameFormulator >= cs.MaxBlocksPerFormulator {
		cs.forwardTop(HeaderHash)
		cs.blocksFromSameFormulator = 0
	}

	phase := cs.largestPhase() + 1
	for _, a := range ctd.CreatedAccountMap {
		if a.Type() == cs.FormulationAccountType {
			acc := a.(*FormulationAccount)
			addr := acc.Address()
			if err := cs.addRank(NewRank(addr, acc.KeyHash, phase, hash.DoubleHash(addr[:]))); err != nil {
				return nil, err
			}
		}
	}
	for _, acc := range ctd.DeletedAccountMap {
		if acc.Type() == cs.FormulationAccountType {
			cs.removeRank(acc.Address())
		}
	}

	SaveData, err := cs.buildSaveData()
	if err != nil {
		return nil, err
	}
	return SaveData, nil
}

func (cs *Consensus) buildSaveData() ([]byte, error) {
	SaveData := []byte{}
	{
		var buffer bytes.Buffer
		if _, err := util.WriteUint64(&buffer, cs.height); err != nil {
			return nil, err
		}
		if _, err := util.WriteUint32(&buffer, cs.MaxBlocksPerFormulator); err != nil {
			return nil, err
		}
		if _, err := util.WriteUint32(&buffer, cs.blocksFromSameFormulator); err != nil {
			return nil, err
		}
		if _, err := util.WriteUint32(&buffer, uint32(len(cs.candidates))); err != nil {
			return nil, err
		} else {
			for _, s := range cs.candidates {
				if _, err := s.WriteTo(&buffer); err != nil {
					return nil, err
				}
			}
		}
		if _, err := util.WriteUint8(&buffer, uint8(len(cs.ObserverKeyMap))); err != nil {
			return nil, err
		}
		for k := range cs.ObserverKeyMap {
			if _, err := k.WriteTo(&buffer); err != nil {
				return nil, err
			}
		}
		SaveData = append(SaveData, buffer.Bytes()...)
	}
	return SaveData, nil
}

// LoadFromSaveData recover the status using the save data
func (cs *Consensus) LoadFromSaveData(SaveData []byte) error {
	cs.Lock()
	defer cs.Unlock()

	r := bytes.NewReader(SaveData)
	if v, _, err := util.ReadUint64(r); err != nil {
		return err
	} else {
		cs.height = v
	}
	if v, _, err := util.ReadUint32(r); err != nil {
		return err
	} else {
		if cs.MaxBlocksPerFormulator != v {
			return ErrInvalidMaxBlocksPerFormulator
		}
	}
	if v, _, err := util.ReadUint32(r); err != nil {
		return err
	} else {
		cs.blocksFromSameFormulator = v
	}
	if Len, _, err := util.ReadUint32(r); err != nil {
		return err
	} else {
		cs.candidates = make([]*Rank, 0, Len)
		cs.rankMap = map[common.Address]*Rank{}
		for i := 0; i < int(Len); i++ {
			s := new(Rank)
			if _, err := s.ReadFrom(r); err != nil {
				return err
			} else {
				cs.candidates = append(cs.candidates, s)
				cs.rankMap[s.Address] = s
			}
		}
	}
	ObserverKeyMap := map[common.PublicHash]bool{}
	if Len, _, err := util.ReadUint8(r); err != nil {
		return err
	} else {
		for i := 0; i < int(Len); i++ {
			var pubhash common.PublicHash
			if _, err := pubhash.ReadFrom(r); err != nil {
				return err
			}
			ObserverKeyMap[pubhash] = true
		}
	}
	cs.ObserverKeyMap = ObserverKeyMap
	return nil
}

func (cs *Consensus) largestPhase() uint32 {
	if len(cs.candidates) == 0 {
		return 0
	}
	return cs.candidates[len(cs.candidates)-1].phase
}

func (cs *Consensus) addRank(s *Rank) error {
	if len(cs.candidates) > 0 {
		if s.Phase() < cs.candidates[0].Phase() {
			return ErrInvalidPhase
		}
	}
	if cs.rankMap[s.Address] != nil {
		return ErrExistAddress
	}
	cs.candidates = InsertRankToList(cs.candidates, s)
	cs.rankMap[s.Address] = s
	return nil
}

func (cs *Consensus) removeRank(addr common.Address) {
	if _, has := cs.rankMap[addr]; has {
		delete(cs.rankMap, addr)
		candidates := make([]*Rank, 0, len(cs.candidates))
		for _, s := range cs.candidates {
			if !s.Address.Equal(addr) {
				candidates = append(candidates, s)
			}
		}
	}
}

func (cs *Consensus) forwardCandidates(TimeoutCount int) error {
	if TimeoutCount >= len(cs.candidates) {
		return ErrExceedCandidateCount
	}

	// increase phase
	for i := 0; i < TimeoutCount; i++ {
		m := cs.candidates[0]
		m.SetPhase(m.Phase() + 2)
		idx := sort.Search(len(cs.candidates)-1, func(i int) bool {
			return m.Less(cs.candidates[i+1])
		})
		copy(cs.candidates, cs.candidates[1:idx+1])
		cs.candidates[idx] = m
	}
	return nil
}

func (cs *Consensus) forwardTop(LastTableAppendHash hash.Hash256) {
	// update top phase and hashSpace
	top := cs.candidates[0]
	top.Set(top.Phase()+1, LastTableAppendHash)
	idx := sort.Search(len(cs.candidates)-1, func(i int) bool {
		return top.Less(cs.candidates[i+1])
	})
	copy(cs.candidates, cs.candidates[1:idx+1])
	cs.candidates[idx] = top

	cs.height++
}

// InsertRankToList inserts the rank by the score to the rank list
func InsertRankToList(ranks []*Rank, s *Rank) []*Rank {
	idx := sort.Search(len(ranks), func(i int) bool {
		return s.Less(ranks[i])
	})
	ranks = append(ranks, s)
	copy(ranks[idx+1:], ranks[idx:])
	ranks[idx] = s
	return ranks
}
