package api

import (
	"net/http"

	"pastebin/internal/store"
)

func Retrieve(w http.ResponseWriter, r *http.Request, s *store.Store, id string) {
	WriteError(w, http.StatusNotImplemented, "GET /pastes/{id} not implemented")
}
