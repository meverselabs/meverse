package pof

import (
	"reflect"
	"sort"
	"sync"

	"github.com/fletaio/fleta/encoding"

	"github.com/fletaio/fleta/common"
	"github.com/fletaio/fleta/common/hash"
)

func init() {
	encoding.Register(RankTable{}, func(enc *encoding.Encoder, rv reflect.Value) error {
		item := rv.Interface().(RankTable)
		if err := enc.EncodeUint32(item.height); err != nil {
			return err
		}
		if err := enc.EncodeArrayLen(len(item.candidates)); err != nil {
			return err
		}
		for _, s := range item.candidates {
			if err := enc.Encode(s); err != nil {
				return err
			}
		}
		return nil
	}, func(dec *encoding.Decoder, rv reflect.Value) error {
		item := NewRankTable()
		if v, err := dec.DecodeUint32(); err != nil {
			return err
		} else {
			item.height = v
		}
		Len, err := dec.DecodeArrayLen()
		if err != nil {
			return err
		}
		for i := 0; i < int(Len); i++ {
			s := new(Rank)
			if err := dec.Decode(&s); err != nil {
				return err
			}
			item.candidates = append(item.candidates, s)
			item.rankMap[s.Address] = s
		}
		rv.Set(reflect.ValueOf(item).Elem())
		return nil
	})
}

// RankTable implements the proof of formulator algorithm
type RankTable struct {
	sync.Mutex
	height     uint32
	candidates []*Rank
	rankMap    map[common.Address]*Rank
}

// NewRankTable returns a RankTable
func NewRankTable() *RankTable {
	cs := &RankTable{
		candidates: []*Rank{},
		rankMap:    map[common.Address]*Rank{},
	}
	return cs
}

// CandidateCount returns a count of the rank table
func (cs *RankTable) CandidateCount() int {
	cs.Lock()
	defer cs.Unlock()

	return len(cs.candidates)
}

// TopRank returns the top rank by Timeoutcount
func (cs *RankTable) TopRank(TimeoutCount int) (*Rank, error) {
	cs.Lock()
	defer cs.Unlock()

	if TimeoutCount >= len(cs.candidates) {
		return nil, ErrInsufficientCandidateCount
	}
	return cs.candidates[TimeoutCount].Clone(), nil
}

// TopRankInMap returns the top rank by the given timeout count in the given map
func (cs *RankTable) TopRankInMap(FormulatorMap map[common.Address]bool) (*Rank, int, error) {
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
func (cs *RankTable) RanksInMap(FormulatorMap map[common.Address]bool, Limit int) ([]*Rank, error) {
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
func (cs *RankTable) IsFormulator(Formulator common.Address, Publichash common.PublicHash) bool {
	cs.Lock()
	defer cs.Unlock()

	rank := cs.rankMap[Formulator]
	if rank == nil {
		return false
	}
	if rank.PublicHash != Publichash {
		return false
	}
	return true
}

func (cs *RankTable) largestPhase() uint32 {
	if len(cs.candidates) == 0 {
		return 0
	}
	return cs.candidates[len(cs.candidates)-1].phase
}

func (cs *RankTable) addRank(s *Rank) error {
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

func (cs *RankTable) removeRank(addr common.Address) {
	if _, has := cs.rankMap[addr]; has {
		delete(cs.rankMap, addr)
		candidates := make([]*Rank, 0, len(cs.candidates))
		for _, s := range cs.candidates {
			if s.Address != addr {
				candidates = append(candidates, s)
			}
		}
	}
}

func (cs *RankTable) forwardCandidates(TimeoutCount int) error {
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

func (cs *RankTable) forwardTop(LastTableAppendHash hash.Hash256) {
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
