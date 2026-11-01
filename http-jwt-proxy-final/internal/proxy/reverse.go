// Package proxy configures forwarding to the private legacy service.
package proxy

import (
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// New builds a reverse proxy and removes credentials before forwarding.
func New(target *url.URL) (*httputil.ReverseProxy,error) {
	if target==nil || target.Host=="" || (target.Scheme!="http" && target.Scheme!="https") { return nil,errors.New("upstream must be an absolute HTTP or HTTPS URL") }
	reverseProxy:=httputil.NewSingleHostReverseProxy(target)
	originalDirector:=reverseProxy.Director
	reverseProxy.Director=func(r *http.Request) {
		originalHost:=r.Host; originalDirector(r); r.Host=target.Host
		r.Header.Del("Authorization"); r.Header.Set("X-Forwarded-Host",originalHost)
		if r.TLS!=nil { r.Header.Set("X-Forwarded-Proto","https") } else { r.Header.Set("X-Forwarded-Proto","http") }
	}
	reverseProxy.Transport=&http.Transport{Proxy:http.ProxyFromEnvironment,DialContext:(&net.Dialer{Timeout:5*time.Second,KeepAlive:30*time.Second}).DialContext,TLSHandshakeTimeout:5*time.Second,ResponseHeaderTimeout:30*time.Second,ExpectContinueTimeout:1*time.Second}
	reverseProxy.ErrorHandler=func(w http.ResponseWriter,_ *http.Request,_ error) { w.Header().Set("Content-Type","application/json"); w.WriteHeader(http.StatusBadGateway); _,_=w.Write([]byte(`{"error":"upstream unavailable"}`+"\n")) }
	return reverseProxy,nil
}
