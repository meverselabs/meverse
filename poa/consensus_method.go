package poa

import (
	"bytes"

	"github.com/fletaio/fleta/encoding"
)

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
	if err := enc.Encode(cs.authorityKey); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
