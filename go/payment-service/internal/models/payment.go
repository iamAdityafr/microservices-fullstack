package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Payment struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrderID       string             `bson:"order_id"       json:"order_id"       validate:"required"`
	UserID        string             `bson:"user_id"        json:"user_id"        validate:"required"`
	Amount        int64              `bson:"amount"         json:"amount"         validate:"gt=0"`
	Currency      string             `bson:"currency"       json:"currency"       validate:"required,len=3"`
	Status        PaymentStatus      `bson:"status"         json:"status"`
	Provider      string             `bson:"provider"       json:"provider"       validate:"required"`
	ProviderRef   string             `bson:"provider_ref"   json:"provider_ref,omitempty"`
	FailureReason string             `bson:"failure_reason" json:"failure_reason,omitempty"`
	CreatedAt     time.Time          `bson:"created_at"     json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"     json:"updated_at"`
}

type PaymentStatus string

const (
	StatusSucceeded PaymentStatus = "succeeded"
	StatusPending   PaymentStatus = "pending"
	StatusFailed    PaymentStatus = "failed"
)

type WebHookEvent struct {
	Type       WebHookEventType
	PaymentID  string
	OrderID    string
	UserID     string
	Amount     int64
	Currency   string
	Status     PaymentStatus
	FailReason string
}

type WebHookEventType string

const (
	EventPaymentSucceeded WebHookEventType = "payment.succeeded"
	EventPaymentFailed    WebHookEventType = "payment.failed"
	EventPaymentRefunded  WebHookEventType = "payment.refunded"
)
