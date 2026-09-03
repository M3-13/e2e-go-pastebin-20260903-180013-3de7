package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"pastebin/internal/store"
)

var hexID16 = regexp.MustCompile(`^[0-9a-f]{16}$`)

func newCreateRecorder(body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/pastes", strings.NewReader(body))
	rec := httptest.NewRecorder()
	Create(rec, req, store.New())
	return rec
}

func TestCreateReturns201WithHexIDAndNoContent(t *testing.T) {
	rec := newCreateRecorder(`{"content":"hallo"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}

	id, ok := body["id"].(string)
	if !ok || !hexID16.MatchString(id) {
		t.Fatalf("id = %v, want a 16-char lowercase hex string", body["id"])
	}

	if _, ok := body["content"]; ok {
		t.Fatal("201 response must not contain a content field")
	}

	for _, k := range []string{"id", "language", "created_at", "expires_at"} {
		if _, ok := body[k]; !ok {
			t.Fatalf("201 response missing key %q", k)
		}
	}
}

func TestCreateInvalidJSONReturns400(t *testing.T) {
	rec := newCreateRecorder(`{"content":`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("error response is not valid JSON: %v", err)
	}
	if _, ok := body["error"]; !ok {
		t.Fatalf("error response missing error field: %v", body)
	}
}

func TestCreateMissingContentReturns400(t *testing.T) {
	rec := newCreateRecorder(`{"language":"go"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateEmptyContentReturns400(t *testing.T) {
	rec := newCreateRecorder(`{"content":""}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateNegativeExpiresReturns400(t *testing.T) {
	rec := newCreateRecorder(`{"content":"hallo","expires_in_seconds":-5}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestCreateBodyOverLimitReturns413(t *testing.T) {
	body := `{"content":"` + strings.Repeat("a", maxBodyBytes+1) + `"}`
	rec := newCreateRecorder(body)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}

func TestCreateDefaultsExpiresTo86400(t *testing.T) {
	rec := newCreateRecorder(`{"content":"hallo"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	createdAt, ok := body["created_at"].(string)
	if !ok {
		t.Fatalf("created_at = %v, want a string", body["created_at"])
	}
	expiresAt, ok := body["expires_at"].(string)
	if !ok {
		t.Fatalf("expires_at = %v, want a string", body["expires_at"])
	}
	if createdAt == "" || expiresAt == "" {
		t.Fatal("timestamps must be set")
	}
}
