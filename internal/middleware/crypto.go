package middleware

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/sshirox/isaac/internal/crypto"
)

type SignValidator struct {
	encoder *crypto.Encoder
}

func NewSignValidator(enc *crypto.Encoder) *SignValidator {
	return &SignValidator{
		encoder: enc,
	}
}

func (s *SignValidator) Validate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s.encoder.IsEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		sign := r.Header.Get(crypto.SignHeader)
		if len(sign) == 0 {
			slog.Info("signature required")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			slog.Error("read request body", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = r.Body.Close()

		body := buf.Bytes()
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		isValid, respSign := s.encoder.Validate(buf.Bytes(), sign)
		if !isValid {
			slog.Info("signature is invalid")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)

		w.Header().Set(crypto.SignHeader, respSign)
	})
}
