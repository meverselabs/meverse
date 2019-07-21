package hash

import (
	"crypto/sha256"
)

// Hash returns the Hash256 value of the data
func Hash(data []byte) Hash256 {
	h := sha256.New()
	if _, err := h.Write(data); err != nil {
		panic(err)
	}
	bs := h.Sum(nil)
	var hash Hash256
	copy(hash[:], bs)
	return hash
}

// DoubleHash returns the result of Hash(Hash(data))
func DoubleHash(data []byte) Hash256 {
	h1 := Hash(data)
	return Hash(h1[:])
}

// Hashes returns the result of Hash(h1+'h'+...)
func Hashes(hs ...Hash256) Hash256 {
	data := make([]byte, (Hash256Size+1)*len(hs)-1)
	idx := 0
	for i, h := range hs {
		copy(data, h[:])
		idx += Hash256Size
		if i < len(hs)-1 {
			data[idx] = 'h'
			idx++
		}
	}
	return Hash(data)
}
