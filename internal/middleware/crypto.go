package middleware

import (
	"bytes"
	"github.com/sshirox/isaac/internal/crypto"
	"io"
	"log/slog"
	"net/http"
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

		sign := r.Header.Get("HashSHA256")
		if len(sign) == 0 {
			slog.Error("validate signature", "empty 'HashSHA256' header")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var buf bytes.Buffer
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			slog.Error("read request body", "err", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		_ = r.Body.Close()

		body := buf.Bytes()
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		isValid, respSign := s.encoder.Validate(buf.Bytes(), sign)
		if !isValid {
			slog.Error("validate signature", "not valid")
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)

		w.Header().Set("HashSHA256", respSign)
	})
}
