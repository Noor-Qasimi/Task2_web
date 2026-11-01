package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var testSecret = []byte("0123456789abcdef0123456789abcdef")

func TestMiddleware(t *testing.T) {
	now := time.Now()
	validClaims := func() jwt.RegisteredClaims {
		return jwt.RegisteredClaims{
			Issuer:    "http-hands-on",
			Subject:   "alice",
			Audience:  jwt.ClaimStrings{"legacy-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(10 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now.Add(-time.Minute)),
		}
	}

	expired := validClaims()
	expired.ExpiresAt = jwt.NewNumericDate(now.Add(-10 * time.Minute))
	wrongIssuer := validClaims()
	wrongIssuer.Issuer = "attacker"
	wrongAudience := validClaims()
	wrongAudience.Audience = jwt.ClaimStrings{"another-service"}
	missingExpiration := validClaims()
	missingExpiration.ExpiresAt = nil

	tests := []struct {
		name       string
		header     string
		wantStatus int
	}{
		{name: "valid token", header: "Bearer " + sign(t, testSecret, validClaims()), wantStatus: http.StatusNoContent},
		{name: "missing header", wantStatus: http.StatusUnauthorized},
		{name: "wrong scheme", header: "Basic abc", wantStatus: http.StatusUnauthorized},
		{name: "malformed token", header: "Bearer not-a-jwt", wantStatus: http.StatusUnauthorized},
		{name: "wrong signature", header: "Bearer " + sign(t, []byte("another-secret-that-is-long-enough"), validClaims()), wantStatus: http.StatusUnauthorized},
		{name: "expired token", header: "Bearer " + sign(t, testSecret, expired), wantStatus: http.StatusUnauthorized},
		{name: "wrong issuer", header: "Bearer " + sign(t, testSecret, wrongIssuer), wantStatus: http.StatusUnauthorized},
		{name: "wrong audience", header: "Bearer " + sign(t, testSecret, wrongAudience), wantStatus: http.StatusUnauthorized},
		{name: "missing expiration", header: "Bearer " + sign(t, testSecret, missingExpiration), wantStatus: http.StatusUnauthorized},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				if got := r.Header.Get("X-Authenticated-Subject"); got != "alice" {
					t.Errorf("verified subject = %q, want alice", got)
				}
				w.WriteHeader(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "http://edge.example/get", nil)
			if test.header != "" {
				req.Header.Set("Authorization", test.header)
			}
			// A client must not be able to spoof this identity header.
			req.Header.Set("X-Authenticated-Subject", "attacker")
			response := httptest.NewRecorder()

			New(testSecret, "http-hands-on", "legacy-api").Middleware(next).ServeHTTP(response, req)
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", response.Code, test.wantStatus, response.Body.String())
			}
			if nextCalled != (test.wantStatus == http.StatusNoContent) {
				t.Fatalf("next handler called = %v", nextCalled)
			}
			if test.wantStatus == http.StatusUnauthorized && response.Header().Get("WWW-Authenticate") == "" {
				t.Error("401 response is missing WWW-Authenticate")
			}
		})
	}
}

func sign(t *testing.T, secret []byte, claims jwt.RegisteredClaims) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}
	return token
}
