package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	UpdatedAt time.Time `json:"updatedAt"`
}
type TaskRepo interface {
	Create(title string) (Task, error)
	Get(id string) (Task, bool)
	List(done bool) []Task
	SetDone(id string, done bool) (Task, error)
}
type Clock interface{ Now() time.Time }

var (
	ErrNotFound     = errors.New("task not found")
	ErrInvalidTitle = errors.New("invalid title")
)

type inMemoryTaskRepo struct {
	mu    sync.RWMutex
	clock Clock
	seq   uint64
	tasks map[string]Task
}

func NewInMemoryTaskRepo(clock Clock) TaskRepo {
	return &inMemoryTaskRepo{
		clock: clock,
		seq:   0,
		tasks: make(map[string]Task),
	}
}
func (r *inMemoryTaskRepo) Create(title string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	res_title := strings.TrimSpace(title)

	if len(res_title) == 0 {
		return Task{}, ErrInvalidTitle
	}

	r.seq++

	safe_task := Task{
		ID:        strconv.FormatUint(r.seq, 10),
		Title:     res_title,
		Done:      false,
		UpdatedAt: r.clock.Now(),
	}

	r.tasks[safe_task.ID] = safe_task

	return safe_task, nil
}
func (r *inMemoryTaskRepo) Get(id string) (Task, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]

	if !ok {
		return Task{}, false
	}

	return task, true
}

func (r *inMemoryTaskRepo) List(done bool) []Task {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := []Task{}

	for _, task := range r.tasks {
		if task.Done == done {
			res = append(res, task)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		if !res[i].UpdatedAt.Equal(res[j].UpdatedAt) {
			return res[i].UpdatedAt.After(res[j].UpdatedAt)
		}
		return res[i].ID < res[j].ID
	})
	return res

}
func (r *inMemoryTaskRepo) SetDone(id string, done bool) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]

	if !ok {
		return Task{}, ErrNotFound
	}

	if task.Done != done {
		task.Done = done
		task.UpdatedAt = r.clock.Now()
		r.tasks[id] = task
	}

	return task, nil
}

type httpHandler struct{ repo TaskRepo }

func NewHTTPHandler(repo TaskRepo) http.Handler {
	return &httpHandler{
		repo: repo,
	}
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	//пути POST /tasks, GET /tasks?done=false, GET /tasks?done=true
	if path == "/tasks" {
		switch r.Method {
		// GET /tasks?done=false, GET /tasks?done=true
		case http.MethodGet:
			h.handleList(w, r)
		// POST /tasks
		case http.MethodPost:
			h.handleCreate(w, r)
		default:
			http.Error(w, "Метод не подреживается", http.StatusMethodNotAllowed)
		}

		return
	}
	//пути GET /tasks/{id}, PATCH /tasks/{id}
	if strings.HasPrefix(path, "/tasks/") {
		id_task := strings.TrimPrefix(path, "/tasks/")

		if id_task == "" || strings.Contains(id_task, "/") {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		switch r.Method {
		case http.MethodGet:
			h.handleGet(w, r, id_task)
		case http.MethodPatch:
			h.handlePatch(w, r, id_task)
		default:
			http.Error(w, "Метод не подреживается", http.StatusMethodNotAllowed)
		}

		return
	}

	http.Error(w, "Not Found", http.StatusNotFound)

}

func (h *httpHandler) handleCreate(w http.ResponseWriter, r *http.Request) {
	var res struct {
		Title string `json:"title"`
	}

	err := decodeStrictJSON(r.Body, &res)
	if err != nil {
		http.Error(w, "Bad request", 400)
		return
	}

	task, err := h.repo.Create(res.Title)
	if errors.Is(err, ErrInvalidTitle) {
		http.Error(w, ErrInvalidTitle.Error(), 400)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeJSON(w, 201, task)

}
func (h *httpHandler) handleGet(w http.ResponseWriter, r *http.Request, id string) {
	task, ok := h.repo.Get(id)

	if !ok {
		http.Error(w, "Нет такого id", 404)
		return
	}

	writeJSON(w, 200, task)
}

func (h *httpHandler) handleList(w http.ResponseWriter, r *http.Request) {
	done_slice := r.URL.Query()["done"]

	var done bool

	if len(done_slice) == 0 {
		done = false
	} else if len(done_slice) > 1 {
		http.Error(w, "Bad request", 400)
		return
	} else {
		res_done := done_slice[0]
		if res_done != "true" && res_done != "false" {
			http.Error(w, "Bad request", 400)
			return
		}
		done = res_done == "true"
	}

	tasks := h.repo.List(done)

	writeJSON(w, 200, tasks)
}
func (h *httpHandler) handlePatch(w http.ResponseWriter, r *http.Request, id string) {
	var res struct {
		Done *bool `json:"done"`
	}

	err := decodeStrictJSON(r.Body, &res)

	if err != nil {
		http.Error(w, "Bad request", 400)
		return
	}

	if res.Done == nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	task, err := h.repo.SetDone(id, *res.Done)

	if errors.Is(err, ErrNotFound) {
		http.Error(w, ErrNotFound.Error(), 404)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	writeJSON(w, 200, task)

}
func decodeStrictJSON(r io.Reader, v any) error {
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(v)
	if err != nil {
		return err
	}

	var extra interface{}
	if err := decoder.Decode(&extra); err == nil {
		return errors.New("trailing data after JSON object")
	} else if err != io.EOF {
		return err
	}
	return nil
}
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	encode := json.NewEncoder(w)
	encode.Encode(v)
}
