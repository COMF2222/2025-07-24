package api

import (
	"2025-07-24/internal/service"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func jsonResponse(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	id, err := h.svc.CreateTask()
	if err != nil {
		if err.Error() == "server busy, max tasks reached" {
			http.Error(w, err.Error(), http.StatusTooManyRequests)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, map[string]string{"id": id}, http.StatusCreated)
}

func (h *Handler) AddLink(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req struct {
		Link string `json:"link"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "недействительный запрос", http.StatusBadRequest)
		return
	}
	if err := h.svc.AddLink(id, req.Link); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse(w, map[string]string{"message": "сслыка добавлена"}, http.StatusOK)
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	task, err := h.svc.GetStatus(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	jsonResponse(w, task, http.StatusOK)
}

func (h *Handler) DownloadArchive(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	file, err := h.svc.GetArchive(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", "attachment; filename="+id+".zip")
	http.ServeFile(w, r, file.Name())
}
