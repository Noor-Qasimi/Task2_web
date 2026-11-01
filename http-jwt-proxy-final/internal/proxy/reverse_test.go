package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)
func TestReverseProxyForwardsRequestWithoutCredentials(t *testing.T) {
	var upstreamHost string
	upstream:=httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter,r *http.Request) {
		if r.Method!=http.MethodPost { t.Errorf("method = %s",r.Method) }
		if r.URL.RequestURI()!="/echo/42?name=ali" { t.Errorf("request URI = %s",r.URL.RequestURI()) }
		if got:=r.Header.Get("Authorization"); got!="" { t.Errorf("Authorization leaked upstream: %q",got) }
		if got:=r.Header.Get("X-Forwarded-Host"); got!="edge.example" { t.Errorf("X-Forwarded-Host = %q",got) }
		if got:=r.Header.Get("X-Forwarded-Proto"); got!="http" { t.Errorf("X-Forwarded-Proto = %q",got) }
		if r.Host!=upstreamHost { t.Errorf("upstream Host = %q, want %q",r.Host,upstreamHost) }
		w.Header().Set("X-Upstream","legacy"); w.WriteHeader(http.StatusCreated); _,_=w.Write([]byte("legacy response"))
	}))
	defer upstream.Close()
	target,_:=url.Parse(upstream.URL); upstreamHost=target.Host
	reverseProxy,err:=New(target); if err!=nil { t.Fatal(err) }
	req:=httptest.NewRequest(http.MethodPost,"http://edge.example/echo/42?name=ali",strings.NewReader("payload"))
	req.Header.Set("Authorization","Bearer secret-token"); req.Header.Set("X-Forwarded-Proto","https")
	response:=httptest.NewRecorder(); reverseProxy.ServeHTTP(response,req)
	if response.Code!=http.StatusCreated { t.Fatalf("status = %d, body=%s",response.Code,response.Body.String()) }
	if response.Header().Get("X-Upstream")!="legacy" { t.Error("upstream response header was not preserved") }
	if response.Body.String()!="legacy response" { t.Errorf("body = %q",response.Body.String()) }
}
func TestReverseProxyReturnsBadGateway(t *testing.T) {
	upstream:=httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter,*http.Request){}))
	target,_:=url.Parse(upstream.URL); upstream.Close()
	reverseProxy,err:=New(target); if err!=nil { t.Fatal(err) }
	response:=httptest.NewRecorder(); reverseProxy.ServeHTTP(response,httptest.NewRequest(http.MethodGet,"http://edge.example/get",nil))
	if response.Code!=http.StatusBadGateway { t.Fatalf("status = %d, want 502",response.Code) }
	if body:=response.Body.String(); !strings.Contains(body,"upstream unavailable") { t.Errorf("unexpected body: %s",body) }
}
func TestNewRejectsInvalidTarget(t *testing.T) {
	for _,value:=range []string{"","localhost:8081","ftp://localhost:8081"} {
		t.Run(fmt.Sprintf("target_%q",value),func(t *testing.T) { target,_:=url.Parse(value); if _,err:=New(target); err==nil { t.Fatalf("New(%q) unexpectedly succeeded",value) } })
	}
}
