package main

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"http-jwt-proxy/internal/auth"
	proxy "http-jwt-proxy/internal/proxy"
)

func main() {
	key := os.Getenv("JWT_SECRET")
	if len(key) < 32 { log.Fatal("JWT_SECRET must contain at least 32 bytes") }
	upstream, err := url.Parse(env("UPSTREAM_URL", "http://legacy:8081"))
	if err != nil { log.Fatal(err) }
	cert, priv := os.Getenv("TLS_CERT_FILE"), os.Getenv("TLS_KEY_FILE")
	if cert == "" || priv == "" { log.Fatal("TLS certificate and key paths are required") }
	forwarder, err := proxy.New(upstream)
	if err != nil { log.Fatal(err) }
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet { w.WriteHeader(http.StatusMethodNotAllowed); return }
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{\"status\":\"ok\"}\n"))
	})
	mux.Handle("/", auth.New([]byte(key), env("JWT_ISSUER", "http-hands-on"), env("JWT_AUDIENCE", "legacy-api")).Middleware(forwarder))
	secure := &http.Server{Addr: env("HTTPS_ADDR", ":8443"), Handler: mux, ReadHeaderTimeout: 5*time.Second}
	redirect := &http.Server{Addr: env("HTTP_ADDR", ":8080"), Handler: http.HandlerFunc(redirectToHTTPS), ReadHeaderTimeout: 5*time.Second}
	errs := make(chan error, 2)
	go func(){ errs <- secure.ListenAndServeTLS(cert, priv) }()
	go func(){ errs <- redirect.ListenAndServe() }()
	log.Printf("proxy servers started: HTTPS %s, HTTP %s", secure.Addr, redirect.Addr)
	if err := <-errs; err != nil && err != http.ErrServerClosed { log.Fatal(err) }
}
func env(k, fallback string) string { if v:=os.Getenv(k); v!="" { return v }; return fallback }

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if h, p, err := net.SplitHostPort(host); err == nil && p == "80" { host = h }
	if host == "" || strings.ContainsAny(host, "\r\n") { http.Error(w, "invalid host", http.StatusBadRequest); return }
	http.Redirect(w, r, "https://"+host+r.URL.RequestURI(), http.StatusPermanentRedirect)
}
