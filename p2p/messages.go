package p2p

import (
	"io"

	"github.com/meverselabs/meverse/common"
	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/meverselabs/meverse/core/types"
)

// message types
var (
	StatusMessageType          = RegisterSerializableType(&StatusMessage{})
	RequestMessageType         = RegisterSerializableType(&RequestMessage{})
	BlockMessageType           = RegisterSerializableType(&BlockMessage{})
	TransactionMessageType     = RegisterSerializableType(&TransactionMessage{})
	PeerListMessageType        = RegisterSerializableType(&PeerListMessage{})
	RequestPeerListMessageType = RegisterSerializableType(&RequestPeerListMessage{})
)

// RequestMessage used to request a chain data to a peer
type RequestMessage struct {
	Height uint32
	Count  uint8
}

func (s *RequestMessage) TypeID() uint32 {
	return RequestMessageType
}

func (s *RequestMessage) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Uint32(w, s.Height); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint8(w, s.Count); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *RequestMessage) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Uint32(r, &s.Height); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint8(r, &s.Count); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}

// StatusMessage used to provide the chain information to a peer
type StatusMessage struct {
	Version  uint16
	Height   uint32
	LastHash hash.Hash256
}

func (s *StatusMessage) TypeID() uint32 {
	return StatusMessageType
}

func (s *StatusMessage) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Uint16(w, s.Version); err != nil {
		return sum, err
	}
	if sum, err := sw.Uint32(w, s.Height); err != nil {
		return sum, err
	}
	if sum, err := sw.Hash256(w, s.LastHash); err != nil {
		return sum, err
	}
	return sw.Sum(), nil
}

func (s *StatusMessage) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if sum, err := sr.Uint16(r, &s.Version); err != nil {
		return sum, err
	}
	if sum, err := sr.Uint32(r, &s.Height); err != nil {
		return sum, err
	}
	if sum, err := sr.Hash256(r, &s.LastHash); err != nil {
		return sum, err
	}
	return sr.Sum(), nil
}

// BlockMessage used to send a chain block to a peer
type BlockMessage struct {
	Blocks []*types.Block
}

func (s *BlockMessage) TypeID() uint32 {
	return BlockMessageType
}

func (s *BlockMessage) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Uint16(w, uint16(len(s.Blocks))); err != nil {
		return sum, err
	}
	for _, v := range s.Blocks {
		if sum, err := sw.WriterTo(w, v); err != nil {
			return sum, err
		}
	}
	return sw.Sum(), nil
}

func (s *BlockMessage) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if Len, sum, err := sr.GetUint16(r); err != nil {
		return sum, err
	} else {
		s.Blocks = make([]*types.Block, 0, Len)
		for i := uint16(0); i < Len; i++ {
			var v types.Block
			if sum, err := sr.ReaderFrom(r, &v); err != nil {
				return sum, err
			}
			s.Blocks = append(s.Blocks, &v)
		}
	}
	return sr.Sum(), nil
}

// TransactionMessage is a message for a transaction
type TransactionMessage struct {
	Txs        []*types.Transaction //MAXLEN : 65535
	Signatures []common.Signature   //MAXLEN : 65535
}

func (s *TransactionMessage) TypeID() uint32 {
	return TransactionMessageType
}

func (s *TransactionMessage) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Uint16(w, uint16(len(s.Txs))); err != nil {
		return sum, err
	}
	for _, v := range s.Txs {
		if sum, err := sw.WriterTo(w, v); err != nil {
			return sum, err
		}
	}
	if sum, err := sw.Uint16(w, uint16(len(s.Signatures))); err != nil {
		return sum, err
	}
	for _, v := range s.Signatures {
		if sum, err := sw.Signature(w, v); err != nil {
			return sum, err
		}
	}
	return sw.Sum(), nil
}

func (s *TransactionMessage) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if Len, sum, err := sr.GetUint16(r); err != nil {
		return sum, err
	} else {
		s.Txs = make([]*types.Transaction, 0, Len)
		for i := uint16(0); i < Len; i++ {
			var v types.Transaction
			if sum, err := sr.ReaderFrom(r, &v); err != nil {
				return sum, err
			}
			s.Txs = append(s.Txs, &v)
		}
	}
	if Len, sum, err := sr.GetUint16(r); err != nil {
		return sum, err
	} else {
		s.Signatures = make([]common.Signature, 0, Len)
		for i := uint16(0); i < Len; i++ {
			var v common.Signature
			if sum, err := sr.Signature(r, &v); err != nil {
				return sum, err
			}
			s.Signatures = append(s.Signatures, v)
		}
	}
	return sr.Sum(), nil
}

// PeerListMessage is a message for a peer list
type PeerListMessage struct {
	Ips   []string
	Hashs []string
}

func (s *PeerListMessage) TypeID() uint32 {
	return PeerListMessageType
}

func (s *PeerListMessage) WriteTo(w io.Writer) (int64, error) {
	sw := bin.NewSumWriter()
	if sum, err := sw.Uint16(w, uint16(len(s.Ips))); err != nil {
		return sum, err
	}
	for _, v := range s.Ips {
		if sum, err := sw.String(w, v); err != nil {
			return sum, err
		}
	}
	if sum, err := sw.Uint16(w, uint16(len(s.Hashs))); err != nil {
		return sum, err
	}
	for _, v := range s.Hashs {
		if sum, err := sw.String(w, v); err != nil {
			return sum, err
		}
	}
	return sw.Sum(), nil
}

func (s *PeerListMessage) ReadFrom(r io.Reader) (int64, error) {
	sr := bin.NewSumReader()
	if Len, sum, err := sr.GetUint16(r); err != nil {
		return sum, err
	} else {
		s.Ips = make([]string, 0, Len)
		for i := uint16(0); i < Len; i++ {
			var v string
			if sum, err := sr.String(r, &v); err != nil {
				return sum, err
			}
			s.Ips = append(s.Ips, v)
		}
	}
	if Len, sum, err := sr.GetUint16(r); err != nil {
		return sum, err
	} else {
		s.Hashs = make([]string, 0, Len)
		for i := uint16(0); i < Len; i++ {
			var v string
			if sum, err := sr.String(r, &v); err != nil {
				return sum, err
			}
			s.Hashs = append(s.Hashs, v)
		}
	}
	return sr.Sum(), nil
}

// RequestPeerListMessage is a request message for a peer list
type RequestPeerListMessage struct {
}

func (s *RequestPeerListMessage) TypeID() uint32 {
	return RequestPeerListMessageType
}

func (s *RequestPeerListMessage) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

func (s *RequestPeerListMessage) ReadFrom(r io.Reader) (int64, error) {
	return 0, nil
}
