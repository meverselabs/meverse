package piledb

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/meverselabs/meverse/common/bin"
	"github.com/meverselabs/meverse/common/hash"
	"github.com/pkg/errors"
)

// Pile proivdes a part of stack like store
type Pile struct {
	sync.Mutex
	file          *os.File
	HeadHeight    uint32
	BeginHeight   uint32
	InitHeight    uint32
	GenHash       hash.Hash256
	InitHash      hash.Hash256
	InitTimestamp uint64
}

// NewPile returns a Pile
func NewPile(path string, GenHash hash.Hash256, InitHash hash.Hash256, InitHeight uint32, InitTimestamp uint64, BaseHeight uint32) (*Pile, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	HeadHeight := BaseHeight
	if BaseHeight < InitHeight {
		HeadHeight = InitHeight
	}

	if fi, err := file.Stat(); err != nil {
		file.Close()
		return nil, errors.WithStack(err)
	} else if fi.Size() < ChunkHeaderSize {
		if err := file.Truncate(ChunkHeaderSize); err != nil {
			file.Close()
			return nil, errors.WithStack(err)
		}
		if _, err := file.Seek(0, 0); err != nil {
			file.Close()
			return nil, errors.WithStack(err)
		}
		meta := make([]byte, ChunkMetaSize)
		if BaseHeight%ChunkUnit != 0 {
			file.Close()
			return nil, errors.WithStack(ErrInvalidChunkBeginHeight)
		}
		copy(meta, bin.Uint32Bytes(HeadHeight))                //HeadHeight (0, 4)
		copy(meta[4:], bin.Uint32Bytes(HeadHeight))            //HeadHeightCheckA (4, 8)
		copy(meta[8:], bin.Uint32Bytes(HeadHeight))            //HeadHeightCheckB (8, 12)
		copy(meta[12:], bin.Uint32Bytes(BaseHeight))           //BeginHeight (12, 16)
		copy(meta[16:], bin.Uint32Bytes(BaseHeight+ChunkUnit)) //EndHeight (16, 20)
		copy(meta[20:], GenHash[:])                            //GenesisHash (20, 52)
		copy(meta[52:], InitHash[:])                           //InitialHash (52, 84)
		copy(meta[84:], bin.Uint32Bytes(InitHeight))           //BeginHeight (84, 88)
		copy(meta[88:], bin.Uint64Bytes(InitTimestamp))        //EndHeight (88, 96)
		if _, err := file.Write(meta); err != nil {
			file.Close()
			return nil, errors.WithStack(err)
		}
	}

	p := &Pile{
		file:        file,
		HeadHeight:  HeadHeight,
		BeginHeight: BaseHeight,
		GenHash:     GenHash,
		InitHash:    InitHash,
	}
	return p, nil
}

// LoadPile loads a pile from the file
func LoadPile(path string) (*Pile, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if _, err := file.Seek(0, 0); err != nil {
		file.Close()
		return nil, errors.WithStack(err)
	}
	meta := make([]byte, ChunkMetaSize)
	if _, err := file.Read(meta); err != nil {
		file.Close()
		return nil, errors.WithStack(err)
	}
	HeadHeight := bin.Uint32(meta)
	HeadHeightCheckA := bin.Uint32(meta[4:])
	HeadHeightCheckB := bin.Uint32(meta[8:])
	BeginHeight := bin.Uint32(meta[12:])
	EndHeight := bin.Uint32(meta[16:])
	var GenHash hash.Hash256
	copy(GenHash[:], meta[20:])
	var InitHash hash.Hash256
	copy(InitHash[:], meta[52:])
	InitHeight := bin.Uint32(meta[84:])
	InitTimestamp := bin.Uint64(meta[88:])
	if BeginHeight%ChunkUnit != 0 {
		file.Close()
		return nil, errors.WithStack(ErrInvalidChunkBeginHeight)
	}
	if BeginHeight+ChunkUnit != EndHeight {
		file.Close()
		return nil, errors.WithStack(ErrInvalidChunkEndHeight)
	}
	if HeadHeight != HeadHeightCheckA || HeadHeightCheckA != HeadHeightCheckB { // crashed when file.Sync()
		if HeadHeightCheckA == HeadHeightCheckB { //crashed at HeadHeight
			if _, err := file.Seek(0, 0); err != nil {
				return nil, errors.WithStack(err)
			}
			if _, err := file.Write(bin.Uint32Bytes(HeadHeightCheckA)); err != nil {
				return nil, errors.WithStack(err)
			}
			if err := file.Sync(); err != nil {
				return nil, errors.WithStack(err)
			}
			HeadHeight = HeadHeightCheckA
		} else if HeadHeight == HeadHeightCheckB+1 { //crashed at HeadHeightCheckA
			if _, err := file.Seek(4, 0); err != nil {
				return nil, errors.WithStack(err)
			}
			if _, err := file.Write(bin.Uint32Bytes(HeadHeight)); err != nil {
				return nil, errors.WithStack(err)
			}
			if _, err := file.Write(bin.Uint32Bytes(HeadHeight)); err != nil {
				return nil, errors.WithStack(err)
			}
			if err := file.Sync(); err != nil {
				return nil, errors.WithStack(err)
			}
			HeadHeightCheckA = HeadHeight
			HeadHeightCheckB = HeadHeight
		} else if HeadHeight == HeadHeightCheckA { //crashed at HeadHeightCheckB
			if _, err := file.Seek(8, 0); err != nil {
				return nil, errors.WithStack(err)
			}
			if _, err := file.Write(bin.Uint32Bytes(HeadHeight)); err != nil {
				return nil, errors.WithStack(err)
			}
			if err := file.Sync(); err != nil {
				return nil, errors.WithStack(err)
			}
			HeadHeightCheckB = HeadHeight
		} else {
			log.Println("PileDB height crashed", HeadHeight, HeadHeightCheckA, HeadHeightCheckB)
			return nil, errors.WithStack(ErrHeightCrashed)
		}
	}
	if true {
		Offset := ChunkHeaderSize
		FromHeight := HeadHeight - BeginHeight
		if FromHeight > 0 {
			if _, err := file.Seek(ChunkMetaSize+(int64(FromHeight)-1)*8, 0); err != nil {
				file.Close()
				return nil, errors.WithStack(err)
			}
			bs := make([]byte, 8)
			if _, err := file.Read(bs); err != nil {
				file.Close()
				return nil, errors.WithStack(err)
			}
			Offset = int64(bin.Uint64(bs))
			if Offset < ChunkHeaderSize {
				Offset = ChunkHeaderSize
			}
		}
		if fi, err := file.Stat(); err != nil {
			file.Close()
			return nil, errors.WithStack(err)
		} else if fi.Size() < Offset {
			return nil, errors.WithStack(ErrInvalidFileSize)
		}
	}
	p := &Pile{
		file:          file,
		HeadHeight:    HeadHeight,
		BeginHeight:   BeginHeight,
		GenHash:       GenHash,
		InitHash:      InitHash,
		InitHeight:    InitHeight,
		InitTimestamp: InitTimestamp,
	}
	return p, nil
}

// Close closes a pile
func (p *Pile) Close() {
	p.Lock()
	defer p.Unlock()

	if p.file != nil {
		p.file.Sync()
		p.file.Close()
		p.file = nil
	}
}

// AppendData pushes data to the top of the pile
func (p *Pile) AppendData(Sync bool, Height uint32, DataHash hash.Hash256, Datas [][]byte) error {
	p.Lock()
	defer p.Unlock()

	if Height != p.HeadHeight+1 {
		return errors.WithStack(ErrInvalidAppendHeight)
	}

	FromHeight := p.HeadHeight - p.BeginHeight

	//get offset
	Offset := ChunkHeaderSize
	if FromHeight > 0 {
		if _, err := p.file.Seek(ChunkMetaSize+(int64(FromHeight)-1)*8, 0); err != nil {
			return errors.WithStack(err)
		}
		bs := make([]byte, 8)
		if _, err := p.file.Read(bs); err != nil {
			return errors.WithStack(err)
		}
		Offset = int64(bin.Uint64(bs))
		if Offset < ChunkHeaderSize {
			Offset = ChunkHeaderSize
		}
	}

	// write data
	if _, err := p.file.Seek(Offset, 0); err != nil {
		return errors.WithStack(err)
	}
	totalLen := int64(32 + 1 + 4*len(Datas))
	if _, err := p.file.Write(DataHash[:]); err != nil {
		return errors.WithStack(err)
	}
	if _, err := p.file.Write([]byte{uint8(len(Datas))}); err != nil {
		return errors.WithStack(err)
	}
	zdatas := make([][]byte, 0, len(Datas))
	for _, v := range Datas {
		var buffer bytes.Buffer
		zw := gzip.NewWriter(&buffer)
		if _, err := zw.Write(v); err != nil {
			return errors.WithStack(err)
		}
		zw.Flush()
		zw.Close()
		zd := buffer.Bytes()
		zdatas = append(zdatas, zd)

		if _, err := p.file.Write(bin.Uint32Bytes(uint32(len(zd)))); err != nil {
			return errors.WithStack(err)
		}
	}
	for _, zd := range zdatas {
		if _, err := p.file.Write(zd); err != nil {
			return errors.WithStack(err)
		}
		totalLen += int64(len(zd))
	}

	// update offset
	if _, err := p.file.Seek(ChunkMetaSize+int64(FromHeight)*8, 0); err != nil {
		return errors.WithStack(err)
	}
	if _, err := p.file.Write(bin.Uint64Bytes(uint64(Offset + totalLen))); err != nil {
		return errors.WithStack(err)
	}

	// update head height
	if _, err := p.file.Seek(0, 0); err != nil {
		return errors.WithStack(err)
	}
	if _, err := p.file.Write(bin.Uint32Bytes(p.HeadHeight + 1)); err != nil {
		return errors.WithStack(err)
	}
	if Sync {
		if err := p.file.Sync(); err != nil {
			return errors.WithStack(err)
		}
	}

	// update head height check A
	if _, err := p.file.Seek(4, 0); err != nil {
		return errors.WithStack(err)
	}
	if _, err := p.file.Write(bin.Uint32Bytes(p.HeadHeight + 1)); err != nil {
		return errors.WithStack(err)
	}
	if Sync {
		if err := p.file.Sync(); err != nil {
			return errors.WithStack(err)
		}
	}

	// update head height check B
	if _, err := p.file.Seek(8, 0); err != nil {
		return errors.WithStack(err)
	}
	if _, err := p.file.Write(bin.Uint32Bytes(p.HeadHeight + 1)); err != nil {
		return errors.WithStack(err)
	}
	if Sync {
		if err := p.file.Sync(); err != nil {
			return errors.WithStack(err)
		}
	}
	p.HeadHeight++
	return nil
}

// GetHash returns a hash value of the height
func (p *Pile) GetHash(Height uint32) (hash.Hash256, error) {
	p.Lock()
	defer p.Unlock()

	FromHeight := Height - p.BeginHeight
	if Height > p.BeginHeight+ChunkUnit {
		return hash.Hash256{}, errors.WithStack(ErrInvalidHeight)
	}

	Offset := ChunkHeaderSize
	if FromHeight > 1 {
		if _, err := p.file.Seek(ChunkMetaSize+(int64(FromHeight)-2)*8, 0); err != nil {
			return hash.Hash256{}, errors.WithStack(err)
		}
		bs := make([]byte, 8)
		if _, err := p.file.Read(bs); err != nil {
			return hash.Hash256{}, errors.WithStack(err)
		}
		Offset = int64(bin.Uint64(bs))
		if Offset < ChunkHeaderSize {
			Offset = ChunkHeaderSize
		}
	}
	if _, err := p.file.Seek(Offset, 0); err != nil {
		return hash.Hash256{}, errors.WithStack(err)
	}
	value := make([]byte, 32)
	if _, err := p.file.Read(value); err != nil {
		return hash.Hash256{}, errors.WithStack(err)
	}
	var h hash.Hash256
	copy(h[:], value)
	return h, nil
}

// GetData returns a data at the index of the height
func (p *Pile) GetData(Height uint32, index int) ([]byte, error) {
	p.Lock()
	defer p.Unlock()

	FromHeight := Height - p.BeginHeight
	if Height > p.BeginHeight+ChunkUnit {
		return nil, errors.WithStack(ErrInvalidHeight)
	}
	if Height > p.HeadHeight {
		return nil, errors.WithStack(ErrInvalidHeight)
	}

	Offset := ChunkHeaderSize
	if FromHeight > 1 {
		if _, err := p.file.Seek(ChunkMetaSize+(int64(FromHeight)-2)*8, 0); err != nil {
			return nil, errors.WithStack(err)
		}
		bs := make([]byte, 8)
		if _, err := p.file.Read(bs); err != nil {
			return nil, errors.WithStack(err)
		}
		Offset = int64(bin.Uint64(bs))
		if Offset < ChunkHeaderSize {
			Offset = ChunkHeaderSize
		}
	}
	if _, err := p.file.Seek(Offset+32, 0); err != nil {
		return nil, errors.WithStack(err)
	}
	lbs := make([]byte, 1)
	if _, err := p.file.Read(lbs); err != nil {
		return nil, errors.WithStack(err)
	}
	if index >= int(lbs[0]) {
		return nil, errors.WithStack(ErrInvalidDataIndex)
	}
	zlbs := make([]byte, 4*lbs[0])
	if _, err := p.file.Read(zlbs); err != nil {
		return nil, errors.WithStack(err)
	}
	zofs := Offset + 32 + 1 + int64(4*lbs[0])
	for i := 0; i < index; i++ {
		zofs += int64(bin.Uint32(zlbs[4*i:]))
	}
	if _, err := p.file.Seek(zofs, 0); err != nil {
		return nil, errors.WithStack(err)
	}

	zsize := bin.Uint32(zlbs[4*index:])
	zd := make([]byte, zsize)
	if _, err := p.file.Read(zd); err != nil {
		return nil, errors.WithStack(err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zd))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	data, err := ioutil.ReadAll(zr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return data, nil
}

// GetDatas returns datas of the height between from and from + count
func (p *Pile) GetDatas(Height uint32, from int, count int) ([]byte, error) {
	p.Lock()
	defer p.Unlock()

	FromHeight := Height - p.BeginHeight
	if Height > p.BeginHeight+ChunkUnit {
		return nil, errors.WithStack(ErrInvalidHeight)
	}
	if Height > p.HeadHeight {
		return nil, errors.WithStack(ErrInvalidHeight)
	}

	Offset := ChunkHeaderSize
	if FromHeight > 1 {
		if _, err := p.file.Seek(ChunkMetaSize+(int64(FromHeight)-2)*8, 0); err != nil {
			return nil, errors.WithStack(err)
		}
		bs := make([]byte, 8)
		if _, err := p.file.Read(bs); err != nil {
			return nil, errors.WithStack(err)
		}
		Offset = int64(bin.Uint64(bs))
		if Offset < ChunkHeaderSize {
			Offset = ChunkHeaderSize
		}
	}
	if _, err := p.file.Seek(Offset+32, 0); err != nil {
		return nil, errors.WithStack(err)
	}
	lbs := make([]byte, 1)
	if _, err := p.file.Read(lbs); err != nil {
		return nil, errors.WithStack(err)
	}
	if from+count > int(lbs[0]) {
		return nil, errors.WithStack(ErrInvalidDataIndex)
	}
	zlbs := make([]byte, 4*lbs[0])
	if _, err := p.file.Read(zlbs); err != nil {
		return nil, errors.WithStack(err)
	}
	zofs := Offset + 32 + 1 + int64(4*lbs[0])
	for i := 0; i < from; i++ {
		zofs += int64(bin.Uint32(zlbs[4*i:]))
	}
	if _, err := p.file.Seek(zofs, 0); err != nil {
		return nil, errors.WithStack(err)
	}
	var buffer bytes.Buffer
	for i := 0; i < count; i++ {
		zsize := bin.Uint32(zlbs[4*(from+i):])
		zd := make([]byte, zsize)
		if _, err := p.file.Read(zd); err != nil {
			return nil, errors.WithStack(err)
		}
		zr, err := gzip.NewReader(bytes.NewReader(zd))
		if err != nil {
			return nil, errors.WithStack(err)
		}
		data, err := ioutil.ReadAll(zr)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		buffer.Write(data)
	}
	return buffer.Bytes(), nil
}
