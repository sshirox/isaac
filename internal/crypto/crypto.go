package crypto

import (
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"github.com/pkg/errors"
	"os"
)

const (
	SignHeader = "HashSHA256"
)

type Encoder struct {
	key       []byte
	isEnabled bool
}

// NewEncoder create new encoder instance
func NewEncoder(key string) *Encoder {
	return &Encoder{
		key:       []byte(key),
		isEnabled: len(key) > 0,
	}
}

// Encode encode passed data
func (e Encoder) Encode(data []byte) string {
	if !e.isEnabled {
		return ""
	}

	return hex.EncodeToString(e.sign(data))
}

// Validate validate passed data by sign
func (e Encoder) Validate(data []byte, sign string) (bool, string) {
	if !e.isEnabled {
		return false, ""
	}

	s, err := hex.DecodeString(sign)
	if err != nil {
		return false, ""
	}

	signedData := e.sign(data)

	return hmac.Equal(signedData, s), hex.EncodeToString(signedData)
}

func (e Encoder) IsEnabled() bool {
	return e.isEnabled
}

func (e Encoder) sign(data []byte) []byte {
	hash := hmac.New(sha256.New, e.key)
	hash.Write(data)

	return hash.Sum(nil)
}

func ReadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	dec, _ := pem.Decode(data)
	if dec == nil || dec.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("invalid private key file")
	}

	return x509.ParsePKCS1PrivateKey(dec.Bytes)
}

func ReadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	dec, _ := pem.Decode(data)
	if dec == nil || dec.Type != "PUBLIC KEY" {
		return nil, errors.New("invalid public key file")
	}

	key, err := x509.ParsePKIXPublicKey(dec.Bytes)
	if err != nil {
		return nil, err
	}

	pkey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("invalid public key")
	}

	return pkey, nil
}
