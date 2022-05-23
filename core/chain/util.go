package chain

import (
	"bytes"

	"github.com/meverselabs/meverse/common/hash"
	"github.com/pkg/errors"
)

const hashPerLevel = 16
const levelHashAppender = "fletablockchain"

// hash16 returns Hash(x1,'f',x2,'l',...,x16)
func hash16(hashes []hash.Hash256) (hash.Hash256, error) {
	if len(hashes) > hashPerLevel {
		return hash.Hash256{}, errors.WithStack(ErrExceedHashCount)
	}

	var buffer bytes.Buffer
	var EmptyHash hash.Hash256
	for i := 0; i < hashPerLevel; i++ {
		if i < len(hashes) {
			buffer.Write(hashes[i][:])
		} else {
			buffer.Write(EmptyHash[:])
		}
		if i < len(levelHashAppender) {
			if err := buffer.WriteByte(byte(levelHashAppender[i])); err != nil {
				return hash.Hash256{}, errors.WithStack(err)
			}
		}
	}
	return hash.DoubleHash(buffer.Bytes()), nil
}

func buildLevel(hashes []hash.Hash256) ([]hash.Hash256, error) {
	LvCnt := len(hashes) / hashPerLevel
	if len(hashes)%hashPerLevel != 0 {
		LvCnt++
	}

	LvHashes := make([]hash.Hash256, LvCnt)
	for i := 0; i < LvCnt; i++ {
		last := (i + 1) * hashPerLevel
		if last > len(hashes) {
			last = len(hashes)
		}
		h, err := hash16(hashes[i*hashPerLevel : last])
		if err != nil {
			return nil, err
		}
		LvHashes[i] = h
	}
	return LvHashes, nil
}

// BuildLevelRoot returns the level root hash
func BuildLevelRoot(hashes []hash.Hash256) (hash.Hash256, error) {
	if len(hashes) > 65536 {
		return hash.Hash256{}, errors.WithStack(ErrExceedHashCount)
	}
	if len(hashes) == 0 {
		return hash.Hash256{}, errors.WithStack(ErrInvalidHashCount)
	}

	lv3, err := buildLevel(hashes)
	if err != nil {
		return hash.Hash256{}, err
	}
	lv2, err := buildLevel(lv3)
	if err != nil {
		return hash.Hash256{}, err
	}
	lv1, err := buildLevel(lv2)
	if err != nil {
		return hash.Hash256{}, err
	}
	h, err := hash16(lv1)
	if err != nil {
		return hash.Hash256{}, err
	}
	return h, nil
}

func isCapitalAndNumber(Name string) bool {
	for i := 0; i < len(Name); i++ {
		c := Name[i]
		if (c < '0' || '9' < c) && (c < 'A' || 'Z' < c) {
			return false
		}
	}
	return true
}

func isAlphabetAndNumber(Name string) bool {
	for i := 0; i < len(Name); i++ {
		c := Name[i]
		if (c < '0' || '9' < c) && (c < 'a' || 'z' < c) && (c < 'A' || 'Z' < c) {
			return false
		}
	}
	return true
}
