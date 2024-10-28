package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

func (e Encoder) Validate(data []byte, signature string) (bool, string) {
	if !e.isEnabled {
		return false, ""
	}

	sign, err := hex.DecodeString(signature)
	if err != nil {
		return false, ""
	}

	signedData := e.sign(data)

	return hmac.Equal(signedData, sign), hex.EncodeToString(signedData)
}

func (e Encoder) IsEnabled() bool {
	return e.isEnabled
}

func (e Encoder) sign(data []byte) []byte {
	h := hmac.New(sha256.New, e.key)
	h.Write(data)

	return h.Sum(nil)
}
