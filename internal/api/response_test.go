package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSONSetsHeaders(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusOK, map[string]string{"status": "ok"})

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
	if no := rec.Header().Get("X-Content-Type-Options"); no != "nosniff" {
		t.Fatalf("X-Content-Type-Options = %q, want nosniff", no)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"status":"ok"}` {
		t.Fatalf("body = %q, want %q", got, `{"status":"ok"}`)
	}
}

func TestWriteErrorBodyOnlyErrorField(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteError(rec, http.StatusBadRequest, "bad input")

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if len(body) != 1 {
		t.Fatalf("error body has %d fields, want exactly 1: %v", len(body), body)
	}
	if body["error"] != "bad input" {
		t.Fatalf("error = %v, want %q", body["error"], "bad input")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}
}

func TestWriteJSONEscapesHTML(t *testing.T) {
	rec := httptest.NewRecorder()
	WriteJSON(rec, http.StatusOK, map[string]string{"content": `<b>&</b>`})

	body := rec.Body.String()
	if strings.Contains(body, "<") || strings.Contains(body, ">") {
		t.Fatalf("body %q contains unescaped HTML", body)
	}
	if !strings.Contains(body, `\u003c`) || !strings.Contains(body, `\u003e`) {
		t.Fatalf("body %q does not contain HTML escapes", body)
	}
}
