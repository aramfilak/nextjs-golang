package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func todoRequest(t *testing.T, handler http.Handler, method, path, body string, status int) *httptest.ResponseRecorder {
	t.Helper()
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(method, path, strings.NewReader(body)))
	if response.Code != status {
		t.Fatalf("%s %s: status = %d, want %d; body = %s", method, path, response.Code, status, response.Body.String())
	}
	return response
}

func TestTodoLifecycle(t *testing.T) {
	handler := newHandler()
	initial := todoRequest(t, handler, http.MethodGet, "/todos", "", http.StatusOK)
	if strings.TrimSpace(initial.Body.String()) != "[]" {
		t.Fatalf("initial list = %s, want []", initial.Body.String())
	}

	created := todoRequest(t, handler, http.MethodPost, "/todos", `{"title":"  Einkaufen  "}`, http.StatusCreated)
	var item todo
	if err := json.Unmarshal(created.Body.Bytes(), &item); err != nil {
		t.Fatal(err)
	}
	if item.ID == "" || item.Title != "Einkaufen" || item.Completed {
		t.Fatalf("unexpected created todo: %+v", item)
	}
	path := "/todos/" + item.ID
	for _, completed := range []bool{true, false} {
		updated := todoRequest(t, handler, http.MethodPatch, path, fmt.Sprintf(`{"completed":%t}`, completed), http.StatusOK)
		if err := json.Unmarshal(updated.Body.Bytes(), &item); err != nil {
			t.Fatal(err)
		}
		if item.Completed != completed {
			t.Fatalf("completed = %t, want %t", item.Completed, completed)
		}
	}
	list := todoRequest(t, handler, http.MethodGet, "/todos", "", http.StatusOK)
	var items []todo
	if err := json.Unmarshal(list.Body.Bytes(), &items); err != nil || len(items) != 1 || items[0] != item {
		t.Fatalf("list = %s, error = %v", list.Body.String(), err)
	}
	todoRequest(t, handler, http.MethodDelete, path, "", http.StatusNoContent)
	remaining := todoRequest(t, handler, http.MethodGet, "/todos", "", http.StatusOK)
	if strings.TrimSpace(remaining.Body.String()) != "[]" {
		t.Fatalf("remaining list = %s, want []", remaining.Body.String())
	}
	todoRequest(t, handler, http.MethodDelete, path, "", http.StatusNotFound)
	todoRequest(t, handler, http.MethodPatch, path, `{"completed":true}`, http.StatusNotFound)
}

func TestTodoValidation(t *testing.T) {
	for _, body := range []string{
		`{}`, `null`, `{"title":"  "}`, `{"title":12}`, `{"title":"valid","unknown":true}`,
		`{"title":"valid"} {}`, `{"title":`, `{"title":"` + strings.Repeat("a", 201) + `"}`,
		`{"title":"` + strings.Repeat("a", 4096) + `"}`,
	} {
		t.Run(body[:min(len(body), 40)], func(t *testing.T) {
			todoRequest(t, newHandler(), http.MethodPost, "/todos", body, http.StatusBadRequest)
		})
	}
	for _, body := range []string{`{}`, `null`, `{"completed":null}`, `{"completed":"true"}`} {
		t.Run("rejects invalid update "+body, func(t *testing.T) {
			todoRequest(t, newHandler(), http.MethodPatch, "/todos/1", body, http.StatusBadRequest)
		})
	}
}

func TestConcurrentTodoCreation(t *testing.T) {
	handler := newHandler()
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(`{"title":"Task"}`)))
			if response.Code != http.StatusCreated {
				t.Errorf("create status = %d, want 201", response.Code)
			}
		}()
	}
	wg.Wait()
	response := todoRequest(t, handler, http.MethodGet, "/todos", "", http.StatusOK)
	var items []todo
	if err := json.Unmarshal(response.Body.Bytes(), &items); err != nil || len(items) != 20 {
		t.Fatalf("todo count = %d, want 20; error = %v", len(items), err)
	}
	ids := make(map[string]bool)
	for _, item := range items {
		if ids[item.ID] {
			t.Fatalf("duplicate ID: %s", item.ID)
		}
		ids[item.ID] = true
	}
}
