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

// Candidates returns a candidates
func (rt *RankTable) Candidates() []*Rank {
	rt.Lock()
	defer rt.Unlock()

	list := []*Rank{}
	for _, c := range rt.candidates {
		list = append(list, c.Clone())
	}
	return list
}

// CandidateCount returns a count of the rank table
func (rt *RankTable) CandidateCount() int {
	rt.Lock()
	defer rt.Unlock()

	return len(rt.candidates)
}

// TopRank returns the top rank by Timeoutcount
func (rt *RankTable) TopRank(TimeoutCount int) (*Rank, error) {
	rt.Lock()
	defer rt.Unlock()

	if TimeoutCount >= len(rt.candidates) {
		return nil, ErrInsufficientCandidateCount
	}
	return rt.candidates[TimeoutCount].Clone(), nil
}

// TopRankInMap returns the top rank by the given timeout count in the given map
func (rt *RankTable) TopRankInMap(FormulatorMap map[common.Address]bool) (*Rank, int, error) {
	rt.Lock()
	defer rt.Unlock()

	if len(FormulatorMap) == 0 {
		return nil, 0, ErrInsufficientCandidateCount
	}
	for i, r := range rt.candidates {
		if FormulatorMap[r.Address] {
			return r, i, nil
		}
	}
	return nil, 0, ErrInsufficientCandidateCount
}

// RanksInMap returns the ranks in the map
func (rt *RankTable) RanksInMap(FormulatorMap map[common.Address]bool, Limit int) ([]*Rank, error) {
	rt.Lock()
	defer rt.Unlock()

	if len(FormulatorMap) == 0 {
		return nil, ErrInsufficientCandidateCount
	}
	if Limit < 1 {
		Limit = 1
	}
	list := make([]*Rank, 0, Limit)
	for _, r := range rt.candidates {
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
func (rt *RankTable) IsFormulator(Formulator common.Address, Publichash common.PublicHash) bool {
	rt.Lock()
	defer rt.Unlock()

	rank := rt.rankMap[Formulator]
	if rank == nil {
		return false
	}
	if rank.PublicHash != Publichash {
		return false
	}
	return true
}

func (rt *RankTable) largestPhase() uint32 {
	if len(rt.candidates) == 0 {
		return 0
	}
	return rt.candidates[len(rt.candidates)-1].phase
}

func (rt *RankTable) addRank(s *Rank) error {
	if len(rt.candidates) > 0 {
		if s.Phase() < rt.candidates[0].Phase() {
			return ErrInvalidPhase
		}
	}
	if rt.rankMap[s.Address] != nil {
		return ErrExistAddress
	}
	rt.candidates = InsertRankToList(rt.candidates, s)
	rt.rankMap[s.Address] = s
	return nil
}

func (rt *RankTable) removeRank(addr common.Address) {
	if _, has := rt.rankMap[addr]; has {
		delete(rt.rankMap, addr)
		candidates := make([]*Rank, 0, len(rt.candidates))
		for _, s := range rt.candidates {
			if s.Address != addr {
				candidates = append(candidates, s)
			}
		}
	}
}

func (rt *RankTable) forwardCandidates(TimeoutCount int) error {
	if TimeoutCount >= len(rt.candidates) {
		return ErrExceedCandidateCount
	}

	// increase phase
	for i := 0; i < TimeoutCount; i++ {
		m := rt.candidates[0]
		m.SetPhase(m.Phase() + 2)
		idx := sort.Search(len(rt.candidates)-1, func(i int) bool {
			return m.Less(rt.candidates[i+1])
		})
		copy(rt.candidates, rt.candidates[1:idx+1])
		rt.candidates[idx] = m
	}
	return nil
}

func (rt *RankTable) forwardTop(LastTableAppendHash hash.Hash256) {
	// update top phase and hashSpace
	top := rt.candidates[0]
	top.Set(top.Phase()+1, LastTableAppendHash)
	idx := sort.Search(len(rt.candidates)-1, func(i int) bool {
		return top.Less(rt.candidates[i+1])
	})
	copy(rt.candidates, rt.candidates[1:idx+1])
	rt.candidates[idx] = top

	rt.height++
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
