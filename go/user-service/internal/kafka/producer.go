package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"user-service/internal/models"

	"github.com/segmentio/kafka-go"
)

type UserProducer struct {
	writer *kafka.Writer
	topic  string
}

func NewUserProducer(brokers []string, topic string) *UserProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
	return &UserProducer{
		writer: writer,
		topic:  topic,
	}
}

func (p *UserProducer) PublishUserCreated(ctx context.Context, event models.UserCreatedEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal user created event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.ID),
		Value: data,
		Headers: []kafka.Header{
			{
				Key:   "event",
				Value: []byte("UserCreated"),
			},
		},
		Time: time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Println("failed to write UserCreated event:", err)
		return err
	}

	log.Println("UserCreated event published:", event.ID)
	return nil
}

func (p *UserProducer) Close() error {
	return p.writer.Close()
}
