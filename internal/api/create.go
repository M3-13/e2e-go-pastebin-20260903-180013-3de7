package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"pastebin/internal/store"
)

const maxBodyBytes = 1 << 20

func Create(w http.ResponseWriter, r *http.Request, s *store.Store) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req struct {
		Content          string `json:"content"`
		Language         string `json:"language"`
		ExpiresInSeconds *int   `json:"expires_in_seconds"`
	}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			WriteError(w, http.StatusRequestEntityTooLarge, "request body too large")
			return
		}
		WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if strings.TrimSpace(req.Content) == "" {
		WriteError(w, http.StatusBadRequest, "content is required")
		return
	}

	expires := 86400 * time.Second
	if req.ExpiresInSeconds != nil {
		if *req.ExpiresInSeconds < 0 {
			WriteError(w, http.StatusBadRequest, "expires_in_seconds must not be negative")
			return
		}
		if *req.ExpiresInSeconds > 0 {
			expires = time.Duration(*req.ExpiresInSeconds) * time.Second
		}
	}

	meta, err := s.Add(req.Content, req.Language, expires)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "could not create paste")
		return
	}

	WriteJSON(w, http.StatusCreated, meta)
}
