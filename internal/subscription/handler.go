package subscription

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/subscribe", h.Subscribe)
	r.Get("/subscriptions", h.GetAll)
	r.Get("/confirm", h.Confirm)
	r.Get("/unsubscribe", h.Unsubscribe)
}

type SubscribeRequest struct {
	Email    string `json:"email"`
	RepoPath string `json:"repo"`
}

func (h *Handler) Subscribe(w http.ResponseWriter, r *http.Request) {
	var req SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.RepoPath == "" {
		http.Error(w, "email and repo are required", http.StatusBadRequest)
		return
	}

	err := h.service.Subscribe(r.Context(), req.Email, req.RepoPath)
	if err != nil {
		// Depending on the error, we might want to return 409 Conflict or 400 Bad Request
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Successfully subscribed. Please check your email for confirmation."))
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "email query parameter is required", http.StatusBadRequest)
		return
	}

	subs, err := h.service.GetAll(r.Context(), email)
	if err != nil {
		http.Error(w, "failed to get subscriptions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(subs); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	if err := h.service.ConfirmSubscription(r.Context(), token); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte("Subscription successfully confirmed!"))
}

func (h *Handler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	if err := h.service.Unsubscribe(r.Context(), token); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Write([]byte("Successfully unsubscribed from repository updates."))
}
