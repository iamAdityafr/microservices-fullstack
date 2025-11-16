package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type CartProducer struct {
	writer *kafka.Writer
}

type CartEvent struct {
	EventType string    `json:"event_type"`
	CartID    string    `json:"cart_id"`
	UserID    string    `json:"user_id"`
	Total     float64   `json:"total"`
	Timestamp time.Time `json:"timestamp"`
}

func NewCartProducer(brokers []string, topic string) *CartProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &CartProducer{writer: writer}
}

func (p *CartProducer) Publish(ctx context.Context, evt CartEvent) error {
	bytes, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(evt.EventType),
		Value: bytes,
		Headers: []kafka.Header{
			{Key: "event", Value: []byte(evt.EventType)},
		},
		Time: time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Printf("failed to publish event %s: %v", evt.EventType, err)
		return err
	}

	log.Printf("published event: %s for cart %s", evt.EventType, evt.CartID)
	return nil
}

func (p *CartProducer) Close() error {
	return p.writer.Close()
}
