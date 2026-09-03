package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pastebin/internal/store"
)

func doList(s *store.Store) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/pastes", nil)
	List(rec, req, s)
	return rec
}

func TestListEmptyStore(t *testing.T) {
	s := store.New()
	rec := doList(s)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var metas []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &metas); err != nil {
		t.Fatalf("body is not a JSON array: %v (body %q)", err, rec.Body.String())
	}
	if len(metas) != 0 {
		t.Fatalf("expected empty list, got %d items: %v", len(metas), metas)
	}
	if got := rec.Body.String(); got != "[]" {
		t.Fatalf("empty store body = %q, want []", got)
	}
}

func TestListReturnsMetaWithoutContent(t *testing.T) {
	s := store.New()
	meta, err := s.Add("hello", "go", time.Hour)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	rec := doList(s)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var metas []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &metas); err != nil {
		t.Fatalf("body is not a JSON array: %v", err)
	}
	if len(metas) != 1 {
		t.Fatalf("expected 1 item, got %d: %v", len(metas), metas)
	}

	item := metas[0]
	for _, field := range []string{"id", "language", "created_at", "expires_at"} {
		if _, ok := item[field]; !ok {
			t.Fatalf("item missing field %q: %v", field, item)
		}
	}
	if _, ok := item["content"]; ok {
		t.Fatalf("item must not contain content field: %v", item)
	}

	if item["id"] != meta.ID {
		t.Fatalf("id = %v, want %v", item["id"], meta.ID)
	}
	if item["language"] != "go" {
		t.Fatalf("language = %v, want %q", item["language"], "go")
	}
}

func TestListExcludesExpiredPastes(t *testing.T) {
	s := store.New()
	if _, err := s.Add("alive", "go", time.Hour); err != nil {
		t.Fatalf("Add: %v", err)
	}
	expired, err := s.Add("gone", "go", -time.Second)
	if err != nil {
		t.Fatalf("Add expired: %v", err)
	}

	rec := doList(s)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var metas []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &metas); err != nil {
		t.Fatalf("body is not a JSON array: %v", err)
	}
	if len(metas) != 1 {
		t.Fatalf("expected 1 item (expired excluded), got %d: %v", len(metas), metas)
	}
	if metas[0]["id"] == expired.ID {
		t.Fatalf("expired paste %q appeared in list", expired.ID)
	}

	if s.Delete(expired.ID) {
		t.Fatalf("expired paste %q should have been removed from store by List", expired.ID)
	}
}
