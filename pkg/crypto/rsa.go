package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

func Encrypt(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, data, nil)
}

func Decrypt(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	return privateKey.Decrypt(nil, data, &rsa.OAEPOptions{Hash: crypto.SHA256})
}
