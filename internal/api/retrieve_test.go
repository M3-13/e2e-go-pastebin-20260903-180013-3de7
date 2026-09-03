package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"pastebin/internal/store"
)

func newRetrieveRequest() *http.Request {
	return httptest.NewRequest(http.MethodGet, "/pastes/ignored", nil)
}

func TestRetrieveReturnsFullPaste(t *testing.T) {
	s := store.New()
	meta, err := s.Add("hello world", "go", time.Hour)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	rec := httptest.NewRecorder()
	Retrieve(rec, newRetrieveRequest(), s, meta.ID)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}

	if body["content"] != "hello world" {
		t.Fatalf("content = %v, want %q", body["content"], "hello world")
	}
	if body["language"] != "go" {
		t.Fatalf("language = %v, want %q", body["language"], "go")
	}
	if body["id"] != meta.ID {
		t.Fatalf("id = %v, want %q", body["id"], meta.ID)
	}
	for _, k := range []string{"created_at", "expires_at"} {
		v, ok := body[k]
		if !ok {
			t.Fatalf("response missing key %q", k)
		}
		if s, ok := v.(string); !ok || s == "" {
			t.Fatalf("key %q = %v, want non-empty string", k, v)
		}
	}
}

func TestRetrieveUnknownID(t *testing.T) {
	s := store.New()

	rec := httptest.NewRecorder()
	Retrieve(rec, newRetrieveRequest(), s, "does-not-exist")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatalf("error response missing error field: %v", body)
	}
	if len(body) != 1 {
		t.Fatalf("error body has %d fields, want exactly 1", len(body))
	}
}

func TestRetrieveExpiredRemovesEntry(t *testing.T) {
	s := store.New()
	meta, err := s.Add("expired", "", -time.Hour)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	rec := httptest.NewRecorder()
	Retrieve(rec, newRetrieveRequest(), s, meta.ID)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	if metas := s.List(); len(metas) != 0 {
		t.Fatalf("List after expired retrieve = %d entries, want 0", len(metas))
	}
}

func TestRetrieveEscapesHTMLContent(t *testing.T) {
	s := store.New()
	meta, err := s.Add("<script>&</script>", "html", time.Hour)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	rec := httptest.NewRecorder()
	Retrieve(rec, newRetrieveRequest(), s, meta.ID)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if strings.Contains(body, "<") || strings.Contains(body, ">") || strings.Contains(body, "&") {
		t.Fatalf("body %q contains unescaped HTML", body)
	}
	for _, esc := range []string{`\u003c`, `\u003e`, `\u0026`} {
		if !strings.Contains(body, esc) {
			t.Fatalf("body %q does not contain escape %q", body, esc)
		}
	}
}
