package main

import (
	"net/http/httptest"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("JWT_SECRET","0123456789abcdef0123456789abcdef")
	t.Setenv("UPSTREAM_URL","http://legacy:8081")
	t.Setenv("JWT_ISSUER","test-issuer")
	t.Setenv("JWT_AUDIENCE","test-audience")
	t.Setenv("TLS_CERT_FILE","/certs/tls.crt")
	t.Setenv("TLS_KEY_FILE","/certs/tls.key")
	cfg,err:=loadConfig()
	if err!=nil { t.Fatal(err) }
	if cfg.upstreamURL.String()!="http://legacy:8081" { t.Errorf("upstream = %s",cfg.upstreamURL) }
	if cfg.jwtIssuer!="test-issuer" || cfg.jwtAudience!="test-audience" { t.Error("JWT settings were not loaded") }
	if cfg.listenAddr!=":8443" || cfg.httpRedirectAddr!=":8080" { t.Error("default HTTPS/redirect addresses were not loaded") }
}
func TestLoadConfigRejectsShortSecret(t *testing.T) {
	t.Setenv("JWT_SECRET","short")
	if _,err:=loadConfig(); err==nil { t.Fatal("short secret was accepted") }
}
func TestLoadConfigRequiresTLS(t *testing.T) {
	t.Setenv("JWT_SECRET","0123456789abcdef0123456789abcdef")
	t.Setenv("TLS_CERT_FILE",""); t.Setenv("TLS_KEY_FILE","")
	if _,err:=loadConfig(); err==nil { t.Fatal("missing TLS configuration was accepted") }
}
func TestLoadConfigRequiresBothTLSFiles(t *testing.T) {
	t.Setenv("JWT_SECRET","0123456789abcdef0123456789abcdef")
	t.Setenv("TLS_CERT_FILE","/certs/tls.crt"); t.Setenv("TLS_KEY_FILE","")
	if _,err:=loadConfig(); err==nil { t.Fatal("incomplete TLS configuration was accepted") }
}
func TestRedirectToHTTPS(t *testing.T) {
	req:=httptest.NewRequest("GET","http://localhost:80/get?source=test",nil)
	rec:=httptest.NewRecorder(); redirectToHTTPS(rec,req)
	if rec.Code!=308 { t.Fatalf("status = %d, want 308",rec.Code) }
	if got,want:=rec.Header().Get("Location"),"https://localhost/get?source=test"; got!=want { t.Fatalf("Location = %q, want %q",got,want) }
}
