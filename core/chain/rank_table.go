package chain

import (
	"io"
	"sort"
	"sync"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/pkg/errors"
)

// RankTable implements the proof of Generator algorithm
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
func (rt *RankTable) TopRank(TimeoutCount uint32) (*Rank, error) {
	rt.Lock()
	defer rt.Unlock()

	if int(TimeoutCount) >= len(rt.candidates) {
		return nil, errors.WithStack(ErrInsufficientCandidateCount)
	}
	return rt.candidates[TimeoutCount].Clone(), nil
}

// TopRankInMap returns the top rank by the given timeout count in the given map
func (rt *RankTable) TopGeneratorInMap(GeneratorMap map[common.Address]bool) (common.Address, uint32, error) {
	rt.Lock()
	defer rt.Unlock()

	if len(GeneratorMap) == 0 {
		return common.Address{}, 0, errors.WithStack(ErrInsufficientCandidateCount)
	}
	for i, r := range rt.candidates {
		if GeneratorMap[r.Address] {
			return r.Address, uint32(i), nil
		}
	}
	return common.Address{}, 0, errors.WithStack(ErrInsufficientCandidateCount)
}

// RanksInMap returns the ranks in the map
func (rt *RankTable) GeneratorsInMap(GeneratorMap map[common.Address]bool, Limit int) ([]common.Address, error) {
	rt.Lock()
	defer rt.Unlock()

	if len(GeneratorMap) == 0 {
		return nil, errors.WithStack(ErrInsufficientCandidateCount)
	}
	if Limit < 1 {
		Limit = 1
	}
	list := make([]common.Address, 0, Limit)
	for _, r := range rt.candidates {
		if GeneratorMap[r.Address] {
			list = append(list, r.Address)
			if len(list) >= Limit {
				break
			}
		}
	}
	return list, nil
}

// IsGenerator returns the given information is correct or not
func (rt *RankTable) IsGenerator(addr common.Address) bool {
	rt.Lock()
	defer rt.Unlock()

	rank := rt.rankMap[addr]
	if rank == nil {
		return false
	}
	return true
}

func (rt *RankTable) smallestPhase() uint32 {
	if len(rt.candidates) == 0 {
		return 0
	}
	return rt.candidates[0].phase
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
			return errors.WithStack(ErrInvalidPhase)
		}
	}
	if rt.rankMap[s.Address] != nil {
		return errors.WithStack(ErrExistAddress)
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
		rt.candidates = candidates
	}
}

func (rt *RankTable) forwardCandidates(TimeoutCount int) error {
	if TimeoutCount >= len(rt.candidates) {
		return errors.WithStack(ErrExceedCandidateCount)
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

func (s *RankTable) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Uint32(w, s.height); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, uint32(len(s.candidates))); err != nil {
		return sum, err
	}
	for _, v := range s.candidates {
		if sum, err := sw.WriterTo(w, v); err != nil {
			return sum, err
		}
	}
	return sw.Sum(), nil
}

func (s *RankTable) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Uint32(r, &s.height); err != nil {
		return sum, err
	}
	if Len, sum, err := sr.GetUint32(r); err != nil {
		return sum, err
	} else {
		s.candidates = make([]*Rank, 0, Len)
		s.rankMap = map[common.Address]*Rank{}
		for i := uint32(0); i < Len; i++ {
			var v Rank
			if sum, err := sr.ReaderFrom(r, &v); err != nil {
				return sum, err
			}
			s.candidates = append(s.candidates, &v)
			s.rankMap[v.Address] = &v
		}
	}
	return sr.Sum(), nil
}
