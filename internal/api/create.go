package api

import (
	"net/http"

	"pastebin/internal/store"
)

func Create(w http.ResponseWriter, r *http.Request, s *store.Store) {
	WriteError(w, http.StatusNotImplemented, "POST /pastes not implemented")
}
