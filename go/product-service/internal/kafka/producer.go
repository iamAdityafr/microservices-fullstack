package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"product-service/internal/models"
	"time"

	"github.com/segmentio/kafka-go"
)

type ProductProducer struct {
	writer *kafka.Writer
	topic  string
}
type ProductUpdatedEvent struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Image      string `json:"image"`
	PriceCents int64  `json:"price_cents"`
}

func NewProductProducer(brokers []string, topic string) *ProductProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
	return &ProductProducer{
		writer: writer,
		topic:  topic,
	}
}
func (p *ProductProducer) PublishProductUpdated(ctx context.Context, event models.ProductUpdatedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}
	msg := kafka.Message{
		Key:   []byte(event.ID),
		Value: data,
		Headers: []kafka.Header{
			{
				Key:   "event",
				Value: []byte("Product Updated"),
			},
		},
		Time: time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Println("Failed to write product updated event : ", err)
		return err
	}
	log.Println("product updated successfully: ", event.ID)
	return nil
}
func (p *ProductProducer) Close() error {
	return p.writer.Close()
}

func (p *ProductProducer) PublishProductDelted(ctx context.Context, event models.ProductDeletedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal the event: %w", err)
	}
	msg := kafka.Message{
		Key:   []byte(event.ID),
		Value: data,
		Headers: []kafka.Header{
			{
				Key:   "event",
				Value: []byte("product deleted"),
			},
		},
		Time: time.Now(),
	}
	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Println("Failed to write produc update event: ", err)
		return err
	}
	log.Println("Product deleted successfully: ", event.ID)
	return nil
}
