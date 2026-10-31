package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"http-jwt-proxy/internal/auth"
	proxy "http-jwt-proxy/internal/proxy"
)

func main() {
	key := os.Getenv("JWT_SECRET")
	if len(key) < 32 {
		log.Fatal("JWT_SECRET must contain at least 32 bytes")
	}

	upstream, err := url.Parse(env("UPSTREAM_URL", "http://legacy:8081"))
	if err != nil {
		log.Fatal(err)
	}

	forwarder, err := proxy.New(upstream)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"status\":\"ok\"}\n"))
	})

	// The health endpoint is public; all legacy-service routes require a valid JWT.
	mux.Handle("/", auth.New(
		[]byte(key),
		env("JWT_ISSUER", "http-hands-on"),
		env("JWT_AUDIENCE", "legacy-api"),
	).Middleware(forwarder))

	server := &http.Server{
		Addr:              env("LISTEN_ADDR", ":8080"),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("HTTP JWT reverse proxy listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
