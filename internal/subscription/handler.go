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

type ErrorResponse struct {
	Error string `json:"error"`
}

// Subscribe godoc
// @Summary      Subscribe to repository updates
// @Description  Subscribes a user to receive email notifications for new commits in a specific GitHub repository.
// @Tags         subscription
// @Produce      text/plain
// @Param        SubscribeRequest body SubscribeRequest true "Subscription request"
// @Success      201  {object}  string
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/subscribe [post]
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

// GetAll godoc
// @Summary      Get all subscriptions for an email
// @Description  Retrieves all subscriptions for a given email address.
// @Tags         subscription
// @Produce      json
// @Param        email  query  string  true  "Email address"
// @Success      200  {object}  []SubscriptionDTO
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/subscriptions [get]
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

// Confirm godoc
// @Summary      Confirm subscription
// @Description  Confirms a subscription using a token.
// @Tags         subscription
// @Produce      text/plain
// @Param        token  query  string  true  "Subscription token"
// @Success      200  {object}  string
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/confirm [get]
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

// Unsubscribe godoc
// @Summary      Unsubscribe from repository updates
// @Description  Unsubscribes a user from receiving email notifications for a specific repository.
// @Tags         subscription
// @Produce      text/plain
// @Param        token  query  string  true  "Subscription token"
// @Success      200  {object}  string
// @Failure      400  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/unsubscribe [get]
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
