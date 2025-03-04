package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"github.com/pkg/errors"
)

func EncryptAES(key []byte, plaintext []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(key)
	key = h.Sum(nil)

	aes, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	return ciphertext, nil
}

func DecryptAES(key []byte, ciphertext []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(key)
	key = h.Sum(nil)

	aes, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	gcm, err := cipher.NewGCM(aes)
	if err != nil {
		panic(err)
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err)
	}

	return plaintext, nil
}
