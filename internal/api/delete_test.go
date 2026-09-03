package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pastebin/internal/store"
)

func TestDeleteExistingPasteReturns204AndRemovesIt(t *testing.T) {
	st := store.New()
	meta, err := st.Add("hello world", "plain", time.Hour)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/pastes/"+meta.ID, nil)
	Delete(rec, req, st, meta.ID)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q, want empty", rec.Body.String())
	}

	if _, ok := st.Get(meta.ID); ok {
		t.Fatalf("paste %s still present in store after DELETE", meta.ID)
	}
}

func TestDeleteUnknownPasteReturns404(t *testing.T) {
	st := store.New()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/pastes/0123456789abcdef", nil)
	Delete(rec, req, st, "0123456789abcdef")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
