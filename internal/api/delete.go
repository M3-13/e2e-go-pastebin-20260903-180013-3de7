package api

import (
	"net/http"

	"pastebin/internal/store"
)

func Delete(w http.ResponseWriter, r *http.Request, s *store.Store, id string) {
	if s.Delete(id) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	WriteError(w, http.StatusNotFound, "paste not found")
}
