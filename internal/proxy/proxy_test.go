package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func forward(t *testing.T, trusted []string, remoteAddr string, headers map[string]string) http.Header {
	t.Helper()
	var got http.Header
	service := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got = r.Header.Clone() }))
	defer service.Close()

	target, _ := url.Parse(service.URL)
	rp, err := New(target, "secret", trusted)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.RemoteAddr = remoteAddr
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	req = req.WithContext(context.WithValue(req.Context(), ctxKey{}, "alice"))
	rp.ServeHTTP(httptest.NewRecorder(), req)
	return got
}

func TestTrustedPeerForwardedHeadersAreHonoured(t *testing.T) {
	got := forward(t, []string{"10.0.0.0/8"}, "10.1.2.3:5555", map[string]string{
		"X-Forwarded-For":   "203.0.113.7",
		"X-Forwarded-Proto": "https",
		"X-Forwarded-Host":  "app.example.com",
	})
	if got.Get("X-Forwarded-For") != "203.0.113.7, 10.1.2.3" {
		t.Errorf("unexpected X-Forwarded-For %q", got.Get("X-Forwarded-For"))
	}
	if got.Get("X-Forwarded-Proto") != "https" || got.Get("X-Forwarded-Host") != "app.example.com" {
		t.Errorf("unexpected proto/host %q %q", got.Get("X-Forwarded-Proto"), got.Get("X-Forwarded-Host"))
	}
}

func TestUntrustedPeerForwardedHeadersAreRebuilt(t *testing.T) {
	got := forward(t, []string{"10.0.0.0/8"}, "198.51.100.9:5555", map[string]string{
		"X-Forwarded-For":   "203.0.113.7",
		"X-Forwarded-Proto": "https",
		"X-Forwarded-Host":  "evil.example",
	})
	if got.Get("X-Forwarded-For") != "198.51.100.9" {
		t.Errorf("spoofed X-Forwarded-For must be replaced, got %q", got.Get("X-Forwarded-For"))
	}
	if got.Get("X-Forwarded-Proto") != "http" || got.Get("X-Forwarded-Host") == "evil.example" {
		t.Errorf("spoofed proto/host must be replaced, got %q %q", got.Get("X-Forwarded-Proto"), got.Get("X-Forwarded-Host"))
	}
}

func TestSingleIPAsTrustedProxy(t *testing.T) {
	got := forward(t, []string{"172.18.0.2"}, "172.18.0.2:1", map[string]string{"X-Forwarded-Proto": "https"})
	if got.Get("X-Forwarded-Proto") != "https" {
		t.Errorf("single IP should be trusted, got %q", got.Get("X-Forwarded-Proto"))
	}
}

func TestNewRejectsInvalidTrustedProxy(t *testing.T) {
	target, _ := url.Parse("http://service")
	if _, err := New(target, "secret", []string{"nope"}); err == nil {
		t.Error("expected error")
	}
}

func TestUpstreamFailureIs502(t *testing.T) {
	target, _ := url.Parse("http://127.0.0.1:1")
	rp, _ := New(target, "secret", nil)
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxKey{}, "alice"))
	rec := httptest.NewRecorder()
	rp.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected 502, got %d", rec.Code)
	}
}
