package piledb

// consts
const (
	ChunkUnit       = uint32(172800 * 30)
	ChunkMetaSize   = int64(256)
	ChunkHeaderSize = int64(int64(ChunkUnit)*8 + ChunkMetaSize)
)
