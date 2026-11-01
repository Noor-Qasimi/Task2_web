package main

import "testing"

func TestLoadConfig(t *testing.T) {
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("UPSTREAM_URL", "http://legacy:8081")
	t.Setenv("JWT_ISSUER", "test-issuer")
	t.Setenv("JWT_AUDIENCE", "test-audience")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.upstreamURL.String() != "http://legacy:8081" {
		t.Errorf("upstream = %s", cfg.upstreamURL)
	}
	if cfg.jwtIssuer != "test-issuer" || cfg.jwtAudience != "test-audience" {
		t.Error("JWT settings were not loaded")
	}
	if cfg.listenAddr != ":8080" {
		t.Errorf("listen address = %q, want :8080", cfg.listenAddr)
	}
}

func TestLoadConfigRejectsShortSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "short")
	if _, err := loadConfig(); err == nil {
		t.Fatal("short secret was accepted")
	}
}

func TestLoadConfigRejectsInvalidUpstream(t *testing.T) {
	t.Setenv("JWT_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("UPSTREAM_URL", "legacy:8081")
	if _, err := loadConfig(); err == nil {
		t.Fatal("relative upstream URL was accepted")
	}
}
