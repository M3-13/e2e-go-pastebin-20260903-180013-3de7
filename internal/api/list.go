package api

import (
	"net/http"

	"pastebin/internal/store"
)

func List(w http.ResponseWriter, r *http.Request, s *store.Store) {
	WriteJSON(w, http.StatusOK, s.List())
}
