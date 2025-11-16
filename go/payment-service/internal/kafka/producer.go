package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
)

type PaymentInitiated struct {
	PaymentID string    `json:"payment_id"`
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Amount    int64     `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type PaymentCaptured struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	CapturedAt time.Time `json:"captured_at"`
}

type PaymentFailed struct {
	PaymentID string    `json:"payment_id"`
	OrderID   string    `json:"order_id"`
	Reason    string    `json:"reason"`
	FailedAt  time.Time `json:"failed_at"`
}

type PaymentProducer struct {
	writer *kafka.Writer
}

func NewPaymentProducer(brokers []string, topic string) *PaymentProducer {
	return &PaymentProducer{
		writer: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		},
	}
}

func (p *PaymentProducer) SendPaymentInitiated(ctx context.Context, evt PaymentInitiated) error {
	b, _ := json.Marshal(evt)
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(evt.OrderID),
		Value: b,
		Headers: []kafka.Header{
			{Key: "event", Value: []byte("PaymentInitiated")},
		},
	})
}

func (p *PaymentProducer) SendPaymentCaptured(ctx context.Context, evt PaymentCaptured) error {
	b, _ := json.Marshal(evt)
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(evt.OrderID),
		Value: b,
		Headers: []kafka.Header{
			{Key: "event", Value: []byte("PaymentCaptured")},
		},
	})
}

func (p *PaymentProducer) SendPaymentFailed(ctx context.Context, evt PaymentFailed) error {
	b, _ := json.Marshal(evt)
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(evt.OrderID),
		Value: b,
		Headers: []kafka.Header{
			{Key: "event", Value: []byte("PaymentFailed")},
		},
	})
}
