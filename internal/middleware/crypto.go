package middleware

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"log/slog"
	"net"
	"net/http"

	"github.com/sshirox/isaac/internal/crypto"
)

type SignValidator struct {
	encoder *crypto.Encoder
}

type CryptoDecoder struct {
	pkey *rsa.PrivateKey
}

func NewSignValidator(enc *crypto.Encoder) *SignValidator {
	return &SignValidator{
		encoder: enc,
	}
}

func NewCryptoDecoder(pkey *rsa.PrivateKey) *CryptoDecoder {
	return &CryptoDecoder{
		pkey: pkey,
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

func (d *CryptoDecoder) Decode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}

		dec, err := rsa.DecryptPKCS1v15(rand.Reader, d.pkey, body)
		if err != nil {
			http.Error(w, "Failed to decrypt data", http.StatusInternalServerError)
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(dec))
		next.ServeHTTP(w, r)
	})
}

// TrustedSubnetMiddleware verifies that the client's IP address is within the trusted subnet.
func TrustedSubnetMiddleware(trustedSubnet string) (func(http.Handler) http.Handler, error) {
	if trustedSubnet == "" {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})
		}, nil
	}

	_, subnet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		slog.Error("Failed to parse trusted subnet",
			slog.String("subnet", trustedSubnet),
			slog.Any("error", err),
		)
		return nil, err
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			ipStr := r.Header.Get("X-Real-IP")
			if ipStr == "" {
				slog.WarnContext(ctx, "X-Real-IP header not found")
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(ipStr)
			if ip == nil || !subnet.Contains(ip) {
				slog.WarnContext(ctx, "Invalid X-Real-IP header",
					slog.String("IP", ipStr),
					slog.String("subnet", trustedSubnet),
				)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}, nil
}
