package main

import (
	"log"
	"net/http"
	"os"

	"pastebin/internal/api"
	"pastebin/internal/store"
)

func newMux(st *store.Store) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/pastes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			api.Create(w, r, st)
		case http.MethodGet:
			api.List(w, r, st)
		default:
			api.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc("/pastes/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			api.WriteError(w, http.StatusNotFound, "not found")
			return
		}
		switch r.Method {
		case http.MethodGet:
			api.Retrieve(w, r, st, id)
		case http.MethodDelete:
			api.Delete(w, r, st, id)
		default:
			api.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		api.WriteError(w, http.StatusNotFound, "not found")
	})

	return mux
}

func main() {
	mux := newMux(store.New())

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("pastebin-api listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
