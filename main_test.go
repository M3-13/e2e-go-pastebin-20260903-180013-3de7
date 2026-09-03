package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pastebin/internal/store"
)

func TestHealthzReturnsOK(t *testing.T) {
	mux := newMux(store.New())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	if no := rec.Header().Get("X-Content-Type-Options"); no != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", no)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"status":"ok"}` {
		t.Fatalf("body = %q, want {\"status\":\"ok\"}", got)
	}
}

func TestPastesWrongMethodReturns405(t *testing.T) {
	mux := newMux(store.New())
	cases := []struct{ method, path string }{
		{http.MethodPut, "/pastes"},
		{http.MethodDelete, "/pastes"},
		{http.MethodPatch, "/pastes/abc"},
		{http.MethodPost, "/pastes/abc"},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(c.method, c.path, nil))
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s %s = %d, want 405", c.method, c.path, rec.Code)
		}
		assertJSONError(t, rec)
	}
}

func TestEmptyIDReturns404(t *testing.T) {
	mux := newMux(store.New())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/pastes/", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	assertJSONError(t, rec)
}

func TestUnknownPathReturnsJSON404(t *testing.T) {
	mux := newMux(store.New())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/no-such-route", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	assertJSONError(t, rec)
}

func TestPastesCollectionRoutesAreWired(t *testing.T) {
	mux := newMux(store.New())
	for _, c := range []struct{ method, path string }{
		{http.MethodPost, "/pastes"},
		{http.MethodGet, "/pastes"},
	} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(c.method, c.path, nil))
		if rec.Code == http.StatusNotFound {
			t.Errorf("%s %s returned 404, route not wired", c.method, c.path)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("%s %s Content-Type = %q, want application/json", c.method, c.path, ct)
		}
	}
}

func assertJSONError(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if no := rec.Header().Get("X-Content-Type-Options"); no != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", no)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Errorf("body %v has no error field", body)
	}
	if len(body) != 1 {
		t.Errorf("body has %d fields, want exactly the error field: %v", len(body), body)
	}
}
