// Command proxy exposes the JWT-authenticated reverse proxy on the public edge.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"http-jwt-proxy/internal/auth"
	reverse "http-jwt-proxy/internal/proxy"
)

type config struct {
	listenAddr  string
	upstreamURL *url.URL
	jwtSecret   []byte
	jwtIssuer   string
	jwtAudience string
}

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	authenticator := auth.New(cfg.jwtSecret, cfg.jwtIssuer, cfg.jwtAudience)
	legacyProxy, err := reverse.New(cfg.upstreamURL)
	if err != nil {
		log.Fatalf("proxy configuration error: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}` + "\n"))
	})
	mux.Handle("/", authenticator.Middleware(legacyProxy))

	server := &http.Server{
		Addr:              cfg.listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	signalContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("HTTP JWT proxy listening on %s", cfg.listenAddr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	case <-signalContext.Done():
		log.Print("shutdown signal received")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			log.Printf("HTTP server shutdown: %v", err)
		}
	}
}

func loadConfig() (config, error) {
	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		return config{}, errors.New("JWT_SECRET must contain at least 32 bytes")
	}

	upstream, err := url.Parse(envOrDefault("UPSTREAM_URL", "http://localhost:8081"))
	if err != nil {
		return config{}, fmt.Errorf("parse UPSTREAM_URL: %w", err)
	}
	if upstream.Host == "" || (upstream.Scheme != "http" && upstream.Scheme != "https") {
		return config{}, errors.New("UPSTREAM_URL must be an absolute HTTP or HTTPS URL")
	}

	return config{
		listenAddr:  envOrDefault("LISTEN_ADDR", ":8080"),
		upstreamURL: upstream,
		jwtSecret:   []byte(secret),
		jwtIssuer:   envOrDefault("JWT_ISSUER", "http-hands-on"),
		jwtAudience: envOrDefault("JWT_AUDIENCE", "legacy-api"),
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
