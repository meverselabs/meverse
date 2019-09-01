package pile

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"os"
	"sync"

	"github.com/fletaio/fleta/common/hash"
	"github.com/fletaio/fleta/common/util"
)

type Pile struct {
	sync.Mutex
	file        *os.File
	HeadHeight  uint32
	BeginHeight uint32
	GenHash     hash.Hash256
}

func NewPile(path string, GenHash hash.Hash256, BaseHeight uint32) (*Pile, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}

	if fi, err := file.Stat(); err != nil {
		file.Close()
		return nil, err
	} else if fi.Size() < ChunkHeaderSize {
		if err := file.Truncate(ChunkHeaderSize); err != nil {
			file.Close()
			return nil, err
		}
		if _, err := file.Seek(0, 0); err != nil {
			file.Close()
			return nil, err
		}
		meta := make([]byte, ChunkMetaSize)
		if BaseHeight%ChunkUnit != 0 {
			file.Close()
			return nil, ErrInvalidChunkBeginHeight
		}
		copy(meta, util.Uint32ToBytes(BaseHeight))               //HeadHeight (0, 4)
		copy(meta[4:], util.Uint32ToBytes(BaseHeight))           //BeginHeight (4, 8)
		copy(meta[8:], util.Uint32ToBytes(BaseHeight+ChunkUnit)) //EndHeight (8, 12)
		copy(meta[12:], GenHash[:])                              //GenesisHash (12, 44)
		if _, err := file.Write(meta); err != nil {
			file.Close()
			return nil, err
		}
	}

	p := &Pile{
		file:        file,
		HeadHeight:  BaseHeight,
		BeginHeight: BaseHeight,
		GenHash:     GenHash,
	}
	return p, nil
}

func LoadPile(path string) (*Pile, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	if _, err := file.Seek(0, 0); err != nil {
		file.Close()
		return nil, err
	}
	meta := make([]byte, ChunkMetaSize)
	if _, err := file.Read(meta); err != nil {
		file.Close()
		return nil, err
	}
	HeadHeight := util.BytesToUint32(meta)
	BeginHeight := util.BytesToUint32(meta[4:])
	EndHeight := util.BytesToUint32(meta[8:])
	var GenHash hash.Hash256
	copy(GenHash[:], meta[12:])
	if BeginHeight%ChunkUnit != 0 {
		file.Close()
		return nil, ErrInvalidChunkBeginHeight
	}
	if BeginHeight+ChunkUnit != EndHeight {
		file.Close()
		return nil, ErrInvalidChunkEndHeight
	}
	if true {
		// truncate unfinished data
		Offset := ChunkHeaderSize
		FromHeight := HeadHeight - BeginHeight
		if FromHeight > 0 {
			if _, err := file.Seek(ChunkMetaSize+(int64(FromHeight)-1)*8, 0); err != nil {
				file.Close()
				return nil, err
			}
			bs := make([]byte, 8)
			if _, err := file.Read(bs); err != nil {
				file.Close()
				return nil, err
			}
			Offset = int64(util.BytesToUint64(bs))
		}
		if fi, err := file.Stat(); err != nil {
			file.Close()
			return nil, err
		} else if fi.Size() < Offset {
			return nil, ErrInvalidFileSize
		}
	}
	p := &Pile{
		file:        file,
		HeadHeight:  HeadHeight,
		BeginHeight: BeginHeight,
		GenHash:     GenHash,
	}
	return p, nil
}

func (p *Pile) Close() {
	p.Lock()
	defer p.Unlock()

	if p.file != nil {
		p.file.Close()
		p.file = nil
	}
}

func (p *Pile) AppendData(Sync bool, Height uint32, DataHash hash.Hash256, Data []byte) error {
	p.Lock()
	defer p.Unlock()

	if Height != p.HeadHeight+1 {
		return ErrInvalidAppendHeight
	}

	FromHeight := p.HeadHeight - p.BeginHeight

	//get offset
	Offset := ChunkHeaderSize
	if FromHeight > 0 {
		if _, err := p.file.Seek(ChunkMetaSize+(int64(FromHeight)-1)*8, 0); err != nil {
			return err
		}
		bs := make([]byte, 8)
		if _, err := p.file.Read(bs); err != nil {
			return err
		}
		Offset = int64(util.BytesToUint64(bs))
	}

	// write data
	if _, err := p.file.Seek(Offset, 0); err != nil {
		return err
	}
	var buffer bytes.Buffer
	buffer.Write(DataHash[:])
	zw := gzip.NewWriter(&buffer)
	if _, err := zw.Write(Data); err != nil {
		return err
	}
	zw.Flush()
	zw.Close()
	data := buffer.Bytes()
	p.file.Write(data)

	// update offset
	if _, err := p.file.Seek(ChunkMetaSize+int64(FromHeight)*8, 0); err != nil {
		return err
	}
	p.file.Write(util.Uint64ToBytes(uint64(Offset + int64(len(data)))))

	// update head height
	if _, err := p.file.Seek(0, 0); err != nil {
		return err
	}
	p.file.Write(util.Uint32ToBytes(p.HeadHeight + 1))

	if Sync {
		if err := p.file.Sync(); err != nil {
			return err
		}
	}
	p.HeadHeight++
	return nil
}

func (p *Pile) GetHash(Height uint32) (hash.Hash256, error) {
	p.Lock()
	defer p.Unlock()

	FromHeight := Height - p.BeginHeight
	if Height > p.BeginHeight+ChunkUnit {
		return hash.Hash256{}, ErrInvalidHeight
	}

	Offset := ChunkHeaderSize
	if FromHeight > 1 {
		if _, err := p.file.Seek(ChunkMetaSize+(int64(FromHeight)-2)*8, 0); err != nil {
			return hash.Hash256{}, err
		}
		bs := make([]byte, 8)
		if _, err := p.file.Read(bs); err != nil {
			return hash.Hash256{}, err
		}
		Offset = int64(util.BytesToUint64(bs))
	}
	if _, err := p.file.Seek(Offset, 0); err != nil {
		return hash.Hash256{}, err
	}
	value := make([]byte, 32)
	if _, err := p.file.Read(value); err != nil {
		return hash.Hash256{}, err
	}
	var h hash.Hash256
	copy(h[:], value)
	return h, nil
}

func (p *Pile) GetData(Height uint32) ([]byte, error) {
	p.Lock()
	defer p.Unlock()

	FromHeight := Height - p.BeginHeight
	if Height > p.BeginHeight+ChunkUnit {
		return nil, ErrInvalidHeight
	}

	Offset := ChunkHeaderSize
	if FromHeight > 1 {
		if _, err := p.file.Seek(ChunkMetaSize+(int64(FromHeight)-2)*8, 0); err != nil {
			return nil, err
		}
		bs := make([]byte, 8)
		if _, err := p.file.Read(bs); err != nil {
			return nil, err
		}
		Offset = int64(util.BytesToUint64(bs))
	}
	var Size int64
	if true {
		if _, err := p.file.Seek(ChunkMetaSize+(int64(FromHeight)-1)*8, 0); err != nil {
			return nil, err
		}
		bs := make([]byte, 8)
		if _, err := p.file.Read(bs); err != nil {
			return nil, err
		}
		Size = int64(util.BytesToUint64(bs)) - Offset
	}

	if _, err := p.file.Seek(Offset+32, 0); err != nil {
		return nil, err
	}
	zd := make([]byte, Size-32)
	if _, err := p.file.Read(zd); err != nil {
		return nil, err
	}
	zr, err := gzip.NewReader(bytes.NewReader(zd))
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(zr)
	if err != nil {
		return nil, err
	}
	return data, nil
}
