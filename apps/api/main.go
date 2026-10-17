package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func newHandler() http.Handler {
	mux := http.NewServeMux()
	store := &todoStore{items: []todo{}}
	mux.HandleFunc("GET /todos", store.list)
	mux.HandleFunc("POST /todos", store.create)
	mux.HandleFunc("PATCH /todos/{id}", store.update)
	mux.HandleFunc("DELETE /todos/{id}", store.delete)
	mux.HandleFunc("GET /hello/{username}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "Hello %s", r.PathValue("username"))
	})
	return mux
}

func main() {
	server := &http.Server{
		Addr:              ":8080",
		Handler:           newHandler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API listening on http://localhost%s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
