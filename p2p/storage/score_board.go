package storage

import "time"

type scoreBoard struct {
	score map[peerGroupType]time.Duration
}

func newScoreBoard() *scoreBoard {
	sb := &scoreBoard{
		score: map[peerGroupType]time.Duration{},
	}

	sb.score[group1] = 0
	sb.score[group2] = 0
	sb.score[group3] = 0

	return sb
}

func (ps *peerStorage) getScoreBoard(score func(string) (time.Duration, bool)) *scoreBoard {
	sb := newScoreBoard()

	ps._getScoreBoard(group1, sb, score)
	ps._getScoreBoard(group2, sb, score)
	ps._getScoreBoard(group3, sb, score)

	return sb
}

func (ps *peerStorage) _getScoreBoard(pt peerGroupType, sb *scoreBoard, score func(string) (time.Duration, bool)) {
	nl := ps.peerGroup[pt]

	var sum time.Duration
	var count int
	for _, pi := range nl {
		if pi == nil {
			break
		}

		if t, has := score(pi.p.Hash); has {
			count++
			sum += t
		}
	}

	if count > 0 {
		sb.score[pt] = time.Duration(sum / time.Duration(count))
	}
}
