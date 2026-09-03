package api

import (
	"net/http"

	"pastebin/internal/store"
)

func Retrieve(w http.ResponseWriter, r *http.Request, s *store.Store, id string) {
	p, ok := s.Get(id)
	if !ok {
		WriteError(w, http.StatusNotFound, "not found")
		return
	}
	WriteJSON(w, http.StatusOK, p)
}
