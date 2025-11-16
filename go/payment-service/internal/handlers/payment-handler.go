package handlers

import (
	"bck/auth/authpb"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"payment-service/internal/database"
	"payment-service/internal/kafka"
	"payment-service/internal/models"
	"payment-service/internal/service"
	"strings"
	"time"
)

type PaymentProvider interface {
	CreatePaymentIntent(ctx context.Context, req service.PaymentCreateRequest) (*service.PaymentCreateResponse, error)
	VerifyWebhook(payload []byte, signature string) (*models.WebHookEvent, error)
}

type PaymentHandler struct {
	repo       database.PaymentRepository
	producer   *kafka.PaymentProducer
	service    PaymentProvider
	authClient authpb.AuthServiceClient
}

func NewPaymentHandler(repo database.PaymentRepository, producer *kafka.PaymentProducer, provider PaymentProvider, authClient authpb.AuthServiceClient) *PaymentHandler {
	return &PaymentHandler{
		repo:       repo,
		producer:   producer,
		service:    provider,
		authClient: authClient,
	}
}

func (h *PaymentHandler) CreateIntent(w http.ResponseWriter, r *http.Request) {
	log.Println("Received CreateIntent request")

	var req struct {
		OrderID  string `json:"order_id"`
		Amount   int64  `json:"amount"`
		Currency string `json:"currency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if strings.TrimSpace(req.OrderID) == "" || req.Amount <= 0 || len(req.Currency) != 3 {
		http.Error(w, "invalid request fields", http.StatusBadRequest)
		return
	}

	cookie, err := r.Cookie("Authorization")
	if err != nil {
		http.Error(w, "cookie not found", http.StatusUnauthorized)
		return
	}

	token := cookie.Value
	authRes, err := h.authClient.ValidateToken(context.Background(), &authpb.ValidateTokenRequest{Token: token})
	if err != nil || !authRes.Valid {
		http.Error(w, "failed to validate token", http.StatusUnauthorized)
		return
	}

	intentResp, err := h.service.CreatePaymentIntent(r.Context(), service.PaymentCreateRequest{
		OrderID:  req.OrderID,
		UserID:   authRes.UserId,
		Amount:   req.Amount,
		Currency: req.Currency,
	})
	if err != nil {
		http.Error(w, "payment provider error: "+err.Error(), http.StatusBadGateway)
		return
	}

	p := &models.Payment{
		OrderID:     req.OrderID,
		UserID:      authRes.UserId,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Status:      models.StatusPending,
		Provider:    "stripe",
		ProviderRef: intentResp.IntentID,
	}

	if err := h.repo.CreatePayment(r.Context(), p); err != nil {
		http.Error(w, "Failed to create payment"+err.Error(), http.StatusInternalServerError)
		return
	}

	initEvent := kafka.PaymentInitiated{
		PaymentID: p.ID.Hex(),
		OrderID:   p.OrderID,
		UserID:    p.UserID,
		Amount:    p.Amount,
		Currency:  p.Currency,
		Status:    string(p.Status),
		CreatedAt: time.Now(),
	}
	if err := h.producer.SendPaymentInitiated(r.Context(), initEvent); err != nil {
		log.Printf("Failed to produce PaymentInitiated event: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"payment_id":    p.ID.Hex(),
		"client_secret": intentResp.ClientSecret,
		"status":        p.Status,
	})
}

func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 2 {
		http.Error(w, "missing orderID", http.StatusBadRequest)
		return
	}
	orderID := parts[1]

	p, err := h.repo.GetPaymentByOrderID(r.Context(), orderID)
	if err != nil {
		http.Error(w, "internal error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if p == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBody = 64 * 1024
	body, err := io.ReadAll(io.LimitReader(r.Body, MaxBody))
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	signature := r.Header.Get("Stripe-Signature")
	event, err := h.service.VerifyWebhook(body, signature)
	if err != nil {
		http.Error(w, "invalid signature: "+err.Error(), http.StatusBadRequest)
		return
	}
	if event == nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.repo.UpdatePaymentStatus(r.Context(), event.OrderID, event.Status, event.PaymentID, event.FailReason); err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	switch event.Type {
	case models.EventPaymentSucceeded:
		kafkaEvent := kafka.PaymentCaptured{
			PaymentID:  event.PaymentID,
			OrderID:    event.OrderID,
			Amount:     event.Amount,
			Currency:   event.Currency,
			CapturedAt: time.Now(),
		}
		_ = h.producer.SendPaymentCaptured(r.Context(), kafkaEvent)

	case models.EventPaymentFailed:
		kafkaEvent := kafka.PaymentFailed{
			PaymentID: event.PaymentID,
			OrderID:   event.OrderID,
			Reason:    event.FailReason,
			FailedAt:  time.Now(),
		}
		_ = h.producer.SendPaymentFailed(r.Context(), kafkaEvent)
	}

	w.WriteHeader(http.StatusOK)
}
