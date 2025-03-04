package util

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"github.com/go-git/go-git/v5/plumbing/hash"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func GenerateRandomRSAKey() (*rsa.PrivateKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	return privKey, errors.Wrapf(err, "failed to generate rsa key")
}

func MarshalPrivateKeyToPem(privKey *rsa.PrivateKey) []byte {
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privKeyPem := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	}
	privKeyPemBytes := pem.EncodeToMemory(&privKeyPem)

	return privKeyPemBytes
}

func GenerateRandomHash(h crypto.Hash) ([]byte, error) {
	hasher := hash.New(h)
	_, err := hasher.Write([]byte(uuid.NewString()))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate random hash")
	}

	return hasher.Sum(nil), nil
}
