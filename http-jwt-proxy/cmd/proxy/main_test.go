package main

import (
	"net/http/httptest"
	"testing"
)

func TestEnvUsesFallbackAndConfiguredValue(t *testing.T) {
	if got:=env("UNSET_PROXY_TEST_VALUE","fallback"); got!="fallback" { t.Fatalf("fallback = %q",got) }
	t.Setenv("PROXY_TEST_VALUE","configured")
	if got:=env("PROXY_TEST_VALUE","fallback"); got!="configured" { t.Fatalf("configured = %q",got) }
}

func TestRedirectToHTTPS(t *testing.T) {
	req:=httptest.NewRequest("GET","http://localhost:80/get?source=test",nil)
	rec:=httptest.NewRecorder()
	redirectToHTTPS(rec,req)
	if rec.Code!=308 { t.Fatalf("status = %d, want 308",rec.Code) }
	if got,want:=rec.Header().Get("Location"),"https://localhost/get?source=test"; got!=want { t.Fatalf("Location = %q, want %q",got,want) }
}

func TestRedirectRejectsInvalidHost(t *testing.T) {
	req:=httptest.NewRequest("GET","http://localhost/get",nil)
	req.Host=""
	rec:=httptest.NewRecorder()
	redirectToHTTPS(rec,req)
	if rec.Code!=400 { t.Fatalf("status = %d, want 400",rec.Code) }
}
