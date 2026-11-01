// Package auth implements JWT authentication for incoming HTTP requests.
package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Authenticator validates HS256 tokens and their registered claims.
type Authenticator struct {
	secret   []byte
	issuer   string
	audience string
}

// New creates an Authenticator with the shared HMAC secret and expected claims.
func New(secret []byte, issuer, audience string) *Authenticator {
	return &Authenticator{secret: secret, issuer: issuer, audience: audience}
}

// Middleware rejects invalid requests before they reach the legacy service.
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			unauthorized(w)
			return
		}

		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(
			parts[1],
			claims,
			func(token *jwt.Token) (any, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
				}
				return a.secret, nil
			},
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
			jwt.WithIssuer(a.issuer),
			jwt.WithAudience(a.audience),
			jwt.WithExpirationRequired(),
			jwt.WithIssuedAt(),
			jwt.WithLeeway(5*time.Second),
		)
		if err != nil || !token.Valid {
			unauthorized(w)
			return
		}

		// Replace any spoofed value and pass only the verified subject upstream.
		r.Header.Del("X-Authenticated-Subject")
		r.Header.Set("X-Authenticated-Subject", claims.Subject)
		next.ServeHTTP(w, r)
	})
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", `Bearer realm="legacy-api"`)
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}` + "\n"))
}
