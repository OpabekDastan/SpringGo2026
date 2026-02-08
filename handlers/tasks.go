package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/OpabekDastan/SpringGo2026.git/models"
)

type TaskHandler struct {
	mu    sync.Mutex
	tasks map[int]models.Task
	next  int
}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{
		tasks: make(map[int]models.Task),
		next:  1,
	}
}

func (h *TaskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		h.get(w, r)
	case http.MethodPost:
		h.create(w, r)
	case http.MethodPatch:
		h.update(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *TaskHandler) get(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	// GET /tasks?id=1
	if idStr != "" {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
			return
		}

		h.mu.Lock()
		task, ok := h.tasks[id]
		h.mu.Unlock()

		if !ok {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "task not found"})
			return
		}

		json.NewEncoder(w).Encode(task)
		return
	}

	// GET /tasks
	h.mu.Lock()
	defer h.mu.Unlock()

	var list []models.Task
	for _, t := range h.tasks {
		list = append(list, t)
	}

	json.NewEncoder(w).Encode(list)
}

func (h *TaskHandler) create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string `json:"title"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Title == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid title"})
		return
	}

	h.mu.Lock()
	task := models.Task{
		ID:    h.next,
		Title: input.Title,
		Done:  false,
	}
	h.tasks[h.next] = task
	h.next++
	h.mu.Unlock()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) update(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid id"})
		return
	}

	var input struct {
		Done *bool `json:"done"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.Done == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid done"})
		return
	}

	h.mu.Lock()
	task, ok := h.tasks[id]
	if !ok {
		h.mu.Unlock()
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "task not found"})
		return
	}

	task.Done = *input.Done
	h.tasks[id] = task
	h.mu.Unlock()

	json.NewEncoder(w).Encode(map[string]bool{"updated": true})
}
