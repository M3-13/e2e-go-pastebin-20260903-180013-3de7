package api

import (
	"net/http"

	"pastebin/internal/store"
)

func Delete(w http.ResponseWriter, r *http.Request, s *store.Store, id string) {
	WriteError(w, http.StatusNotImplemented, "DELETE /pastes/{id} not implemented")
}
