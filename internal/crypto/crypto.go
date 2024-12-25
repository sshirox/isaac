package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

const (
	SignHeader = "HashSHA256"
)

type Encoder struct {
	key       []byte
	isEnabled bool
}

func NewEncoder(key string) *Encoder {
	return &Encoder{
		key:       []byte(key),
		isEnabled: len(key) > 0,
	}
}

func (e Encoder) Encode(data []byte) string {
	if !e.isEnabled {
		return ""
	}

	return hex.EncodeToString(e.sign(data))
}

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
