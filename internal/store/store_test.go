package store

import (
	"encoding/json"
	"regexp"
	"testing"
	"time"
)

var hexID = regexp.MustCompile(`^[0-9a-f]{16}$`)

func TestAddAndGet(t *testing.T) {
	s := New()

	meta, err := s.Add("hello world", "go", time.Hour)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if !hexID.MatchString(meta.ID) {
		t.Fatalf("id %q is not 16 hex chars", meta.ID)
	}
	if meta.Language != "go" {
		t.Fatalf("meta language = %q, want %q", meta.Language, "go")
	}
	if meta.CreatedAt.IsZero() || meta.ExpiresAt.IsZero() {
		t.Fatal("timestamps must be set")
	}

	p, ok := s.Get(meta.ID)
	if !ok {
		t.Fatal("Get of existing paste returned false")
	}
	if p.Content != "hello world" {
		t.Fatalf("content = %q", p.Content)
	}
	if p.Language != "go" {
		t.Fatalf("language = %q", p.Language)
	}
	if p.ID != meta.ID {
		t.Fatalf("id = %q, want %q", p.ID, meta.ID)
	}
}

func TestGetUnknown(t *testing.T) {
	s := New()
	if _, ok := s.Get("does-not-exist"); ok {
		t.Fatal("Get of unknown paste returned true")
	}
}

func TestGetExpiredRemovesEntry(t *testing.T) {
	s := New()
	meta, err := s.Add("expired", "", -time.Hour)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	if _, ok := s.Get(meta.ID); ok {
		t.Fatal("Get of expired paste returned true")
	}

	// Lazy deletion: the expired entry must be gone from the store.
	if metas := s.List(); len(metas) != 0 {
		t.Fatalf("List after expired Get = %d entries, want 0", len(metas))
	}
}

func TestListOmitsContentAndExpired(t *testing.T) {
	s := New()

	live, err := s.Add("secret-live", "go", time.Hour)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if _, err := s.Add("secret-expired", "go", -time.Hour); err != nil {
		t.Fatalf("Add: %v", err)
	}

	metas := s.List()
	if len(metas) != 1 {
		t.Fatalf("List = %d entries, want 1 (expired must be excluded)", len(metas))
	}
	if metas[0].ID != live.ID {
		t.Fatalf("List id = %q, want %q", metas[0].ID, live.ID)
	}
}

func TestDelete(t *testing.T) {
	s := New()
	meta, err := s.Add("to delete", "", time.Hour)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}

	if !s.Delete(meta.ID) {
		t.Fatal("Delete of existing paste returned false")
	}
	if s.Delete(meta.ID) {
		t.Fatal("second Delete returned true, want false")
	}
	if _, ok := s.Get(meta.ID); ok {
		t.Fatal("Get after Delete returned true")
	}
}

func TestDeleteUnknown(t *testing.T) {
	s := New()
	if s.Delete("does-not-exist") {
		t.Fatal("Delete of unknown paste returned true")
	}
}

func TestIDFormat16Hex(t *testing.T) {
	s := New()
	seen := map[string]bool{}
	for i := 0; i < 50; i++ {
		meta, err := s.Add("x", "", time.Hour)
		if err != nil {
			t.Fatalf("Add: %v", err)
		}
		if len(meta.ID) != 16 {
			t.Fatalf("id %q has length %d, want 16", meta.ID, len(meta.ID))
		}
		if !hexID.MatchString(meta.ID) {
			t.Fatalf("id %q is not lowercase hex", meta.ID)
		}
		if seen[meta.ID] {
			t.Fatalf("id %q was generated twice", meta.ID)
		}
		seen[meta.ID] = true
	}
}

func TestPasteMetaMarshalsWithoutContent(t *testing.T) {
	now := time.Now()
	m := PasteMeta{
		ID:        "0123456789abcdef",
		Language:  "go",
		CreatedAt: now,
		ExpiresAt: now.Add(time.Hour),
	}

	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if _, ok := raw["content"]; ok {
		t.Fatal("PasteMeta JSON must not contain a content field")
	}
	for _, k := range []string{"id", "language", "created_at", "expires_at"} {
		if _, ok := raw[k]; !ok {
			t.Fatalf("PasteMeta JSON missing key %q", k)
		}
	}
}
