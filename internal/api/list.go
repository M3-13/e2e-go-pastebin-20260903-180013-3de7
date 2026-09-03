package api

import (
	"net/http"

	"pastebin/internal/store"
)

func List(w http.ResponseWriter, r *http.Request, s *store.Store) {
	WriteError(w, http.StatusNotImplemented, "GET /pastes not implemented")
}
