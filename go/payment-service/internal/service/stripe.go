package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"payment-service/internal/kafka"
	"payment-service/internal/models"
	"time"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/paymentintent"
	"github.com/stripe/stripe-go/v78/webhook"
)

type StripeProvider struct {
	secretKey       string
	webhookSecret   string
	paymentProducer *kafka.PaymentProducer
}

type PaymentCreateRequest struct {
	OrderID  string
	UserID   string
	Amount   int64
	Currency string
}

type PaymentCreateResponse struct {
	IntentID     string
	ClientSecret string
	Status       string
}

type PaymentCapturedEvent struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	CapturedAt time.Time `json:"captured_at"`
}

type PaymentFailedEvent struct {
	PaymentID string    `json:"payment_id"`
	OrderID   string    `json:"order_id"`
	Reason    string    `json:"reason"`
	FailedAt  time.Time `json:"failed_at"`
}

func NewStripeProvider(secretKey, webhookSecret string, paymentProducer *kafka.PaymentProducer) *StripeProvider {
	stripe.Key = secretKey
	return &StripeProvider{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
		paymentProducer: kafka.NewPaymentProducer(
			[]string{"localhost:9092"},
			"payment-events",
		),
	}
}

func (s *StripeProvider) CreatePaymentIntent(ctx context.Context, req PaymentCreateRequest) (*PaymentCreateResponse, error) {
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(req.Amount),
		Currency: stripe.String(req.Currency),
		Metadata: map[string]string{
			"order_id": req.OrderID,
			"user_id":  req.UserID,
		},
	}

	intent, err := paymentintent.New(params)
	if err != nil {
		log.Printf("Stripe create intent failed: %v", err)
		return nil, fmt.Errorf("stripe error: %w", err)
	}

	log.Printf("Stripe PaymentIntent created: %s", intent.ID)

	return &PaymentCreateResponse{
		IntentID:     intent.ID,
		ClientSecret: intent.ClientSecret,
		Status:       string(intent.Status),
	}, nil
}

func (s *StripeProvider) VerifyWebhook(payload []byte, signature string) (*models.WebHookEvent, error) {
	event, err := webhook.ConstructEvent(payload, signature, s.webhookSecret)
	if err != nil {
		log.Printf("Webhook verification failed: %v", err)
		return nil, fmt.Errorf("webhook verification failed: %w", err)
	}

	switch event.Type {
	case "payment_intent.succeeded":
		return s.handlePaymentSucceeded(event.Data.Raw)
	case "payment_intent.payment_failed":
		return s.handlePaymentFailed(event.Data.Raw)
	default:
		log.Printf("Ignored webhook event type: %s", event.Type)
		return nil, nil
	}
}

func (s *StripeProvider) handlePaymentSucceeded(raw json.RawMessage) (*models.WebHookEvent, error) {
	var intent stripe.PaymentIntent
	if err := json.Unmarshal(raw, &intent); err != nil {
		log.Printf("Failed to unmarshal PaymentIntent (succeeded): %v", err)
		return nil, err
	}

	log.Printf("succeeded for PaymentIntent: %s", intent.ID)

	return &models.WebHookEvent{
		Type:      models.EventPaymentSucceeded,
		PaymentID: intent.ID,
		OrderID:   intent.Metadata["order_id"],
		UserID:    intent.Metadata["user_id"],
		Amount:    intent.Amount,
		Currency:  string(intent.Currency),
		Status:    models.StatusSucceeded,
	}, nil
}

func (s *StripeProvider) handlePaymentFailed(raw json.RawMessage) (*models.WebHookEvent, error) {
	var intent stripe.PaymentIntent
	if err := json.Unmarshal(raw, &intent); err != nil {
		log.Printf("Error unmarshal PaymentIntent: %v", err)
		return nil, err
	}

	failReason := ""
	if intent.LastPaymentError != nil {
		failReason = intent.LastPaymentError.Msg
	}
	event := kafka.PaymentFailed{
		PaymentID: intent.ID,
		OrderID:   intent.Metadata["order_id"],
		Reason:    failReason,
		FailedAt:  time.Now(),
	}
	log.Printf("failed for PaymentIntent: %s, reason: %s", intent.ID, failReason)

	if err := s.paymentProducer.SendPaymentFailed(context.Background(), event); err != nil {
		log.Printf("Failed to publish PaymentFailed event: %v", err)
		return nil, err
	}

	return &models.WebHookEvent{
		Type:       models.EventPaymentFailed,
		PaymentID:  intent.ID,
		OrderID:    intent.Metadata["order_id"],
		UserID:     intent.Metadata["user_id"],
		Amount:     intent.Amount,
		Currency:   string(intent.Currency),
		Status:     models.StatusFailed,
		FailReason: failReason,
	}, nil
}
