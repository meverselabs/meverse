package main // import "github.com/fletaio/fleta"

import (
	"log"
	"time"

	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/core/chain"
	"github.com/fletaio/fleta/core/types"
	"github.com/fletaio/fleta/encoding"
)

func main() {
	if err := test(); err != nil {
		panic(err)
	}
}

func test() error {
	st, err := chain.NewStore("./_data", "FLEAT Mainnet", 0x0001, true)
	if err != nil {
		return err
	}

	cs := &Consensus{}
	cn := chain.NewChain(cs, st)
	if err := cn.Init(); err != nil {
		return err
	}
	ctx := types.NewContext(cn.Loader())
	TxHashes := []hash.Hash256{ctx.LastHash()}
	LevelRootHash, err := chain.BuildLevelRoot(TxHashes)
	if err != nil {
		return err
	}
	now := uint64(time.Now().UnixNano())
	if now <= ctx.LastTimestamp() {
		now = ctx.LastTimestamp() + 1
	}

	b := &types.Block{
		Header: types.Header{
			Version:       ctx.Version(),
			Height:        ctx.TargetHeight(),
			PrevHash:      ctx.LastHash(),
			LevelRootHash: LevelRootHash,
			ContextHash:   ctx.Hash(),
			Timestamp:     now,
			ConsensusData:          []byte{0},
		},
	}
	if err := cn.ConnectBlock(b); err != nil {
		return err
	}
	if true {
		b, err := cn.Provider().Block(cn.Provider().Height())
		if err != nil {
			return err
		}
		log.Println(cn.Provider().Height(), b, encoding.Hash(b.Header))
	}
	return nil
}

type Consensus struct {
	*chain.ConsensusBase
	cn *chain.Chain
	ct chain.Committer
}

func (cs *Consensus) Init(cn *chain.Chain, ct chain.Committer) error {
	cs.cn = cn
	cs.ct = ct
	return nil
}
