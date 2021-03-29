package bank

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"encoding/hex"

	"github.com/fletaio/fleta/common/hash"
)

//Cipher cipher data
func Cipher(src []byte, Password string) ([]byte, error) {
	SeedHash := hash.DoubleHash([]byte(Password + "@fleta.bank@seed"))

	hasher := sha512.New()
	hasher.Write(SeedHash[:])
	out := hex.EncodeToString(hasher.Sum(nil))
	newKey, err := hex.DecodeString(out[:64])
	nonce, err := hex.DecodeString(out[64:(64 + 24)])

	block, err := aes.NewCipher(newKey)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	return aesgcm.Seal(nil, nonce, src, []byte("fleta.bank")), nil
}

//Decipher decrypting data
func Decipher(cipherText []byte, Password string) ([]byte, error) {
	SeedHash := hash.DoubleHash([]byte(Password + "@fleta.bank@seed"))

	hasher := sha512.New()
	hasher.Write(SeedHash[:])
	out := hex.EncodeToString(hasher.Sum(nil))
	newKey, err := hex.DecodeString(out[:64])
	if err != nil {
		return nil, err
	}
	nonce, err := hex.DecodeString(out[64:(64 + 24)])
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(newKey)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	output, err := aesgcm.Open(nil, nonce, cipherText, []byte("fleta.bank"))
	if err != nil {
		return nil, err
	}
	return output, nil
}
