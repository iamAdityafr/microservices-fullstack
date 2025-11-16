package kafka

import (
	"context"
	"encoding/json"
	"log"
	"payment-service/internal/database"
	"time"

	"github.com/segmentio/kafka-go"
)

type PaymentConsumer struct {
	reader      *kafka.Reader
	paymentRepo database.PaymentRepository
}

func NewPaymentConsumer(brokers []string, topic, groupId string, repo database.PaymentRepository) *PaymentConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupId,
	})
	return &PaymentConsumer{
		reader:      reader,
		paymentRepo: repo,
	}
}

func (c *PaymentConsumer) Consume(ctx context.Context) error {
	log.Println("Payment cart listening for payment publish")

	for {
		select {
		case <-ctx.Done():
			log.Println("PaymentConsuemr stopping")
			return nil
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				log.Println("Error fetch messaage", err)
				continue
			}

			if err := c.ProcessMessages(ctx, msg); err != nil {
				log.Println("Error processing messages", err)
			} else {
				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					log.Println("Failed to commit message: ", err)
				}
			}
		}
	}
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

func (c *PaymentConsumer) ProcessMessages(ctx context.Context, msg kafka.Message) error {
	var eventType string
	for _, h := range msg.Headers {
		if h.Key == "event" {
			eventType = string(h.Value)
			break
		}
	}
	log.Println("[PaymentConsumer] received event type:", eventType)

	switch eventType {
	case "PaymentCaptured":
		var evt PaymentCaptured
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			return err
		}
		log.Printf("[PaymentCaptured] order: %s", evt.OrderID)
		return nil

	case "PaymentFailed":
		var evt PaymentFailed
		if err := json.Unmarshal(msg.Value, &evt); err != nil {
			return err
		}
		log.Printf("[PaymentFailed] order %s", evt.OrderID)
		return nil

	default:
		log.Println("unknown event type")
		return nil
	}
}
