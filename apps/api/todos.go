package main

import (
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

type todo struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type todoStore struct {
	mu     sync.Mutex
	items  []todo
	nextID uint64
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func readJSON(w http.ResponseWriter, r *http.Request, value any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		http.Error(w, "Expected a single JSON object", http.StatusBadRequest)
		return false
	}
	return true
}

func (s *todoStore) list(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, s.items)
}

func (s *todoStore) create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}
	if !readJSON(w, r, &input) {
		return
	}
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" || utf8.RuneCountInString(input.Title) > 200 {
		http.Error(w, "Title must contain between 1 and 200 characters", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	item := todo{ID: strconv.FormatUint(s.nextID, 10), Title: input.Title}
	s.items = append(s.items, item)
	w.Header().Set("Location", "/todos/"+item.ID)
	writeJSON(w, http.StatusCreated, item)
}

func (s *todoStore) update(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Completed *bool `json:"completed"`
	}
	if !readJSON(w, r, &input) {
		return
	}
	if input.Completed == nil {
		http.Error(w, "Completed is required", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	index := s.index(r.PathValue("id"))
	if index < 0 {
		http.NotFound(w, r)
		return
	}
	s.items[index].Completed = *input.Completed
	writeJSON(w, http.StatusOK, s.items[index])
}

func (s *todoStore) delete(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	index := s.index(r.PathValue("id"))
	if index < 0 {
		http.NotFound(w, r)
		return
	}
	s.items = slices.Delete(s.items, index, index+1)
	w.WriteHeader(http.StatusNoContent)
}

func (s *todoStore) index(id string) int {
	return slices.IndexFunc(s.items, func(item todo) bool { return item.ID == id })
}
