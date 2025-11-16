package kafka

import (
	"cart-service/internal/database"
	"cart-service/internal/models"
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type CartConsumer struct {
	reader   *kafka.Reader
	cartRepo database.CartRepository
}

func NewCartConsumer(brokers []string, topic string, groupID string, repo database.CartRepository) *CartConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: groupID,
	})
	return &CartConsumer{
		reader:   reader,
		cartRepo: repo,
	}
}

func (c *CartConsumer) Consume(ctx context.Context) error {
	log.Println("CartConsumer started ...")

	for {
		select {
		case <-ctx.Done():
			log.Println("CartConsumer graceful shutdown")
			return nil
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				log.Println("Error fetching message:", err)
				continue
			}

			if err := c.ProcessMessage(ctx, msg); err != nil {
				log.Println("Error processing message:", err)
			} else {
				if err := c.reader.CommitMessages(ctx, msg); err != nil {
					log.Println("Couldn't commit message:", err)
				}
			}
		}
	}
}

func (c *CartConsumer) ProcessMessage(ctx context.Context, msg kafka.Message) error {
	eventType := ""
	for _, h := range msg.Headers {
		if strings.ToLower(h.Key) == "event" {
			eventType = string(h.Value)
			break
		}
	}

	log.Printf("[CartConsumer] Received event type: %s", eventType)

	if eventType != "Product Updated" {
		log.Printf("Ignore event type: %s", eventType)
		return nil
	}

	var event ProductUpdatedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return err
	}

	return c.handleProductUpdated(ctx, event)
}

type ProductUpdatedEvent struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Image      string `json:"image"`
	PriceCents int64  `json:"price_cents"`
}

func (c *CartConsumer) handleProductUpdated(ctx context.Context, event ProductUpdatedEvent) error {
	log.Printf("Listening ProductUpdated event for productID=%s", event.ID)

	update := models.CartItem{
		ProductID:  event.ID,
		Name:       event.Name,
		Image:      event.Image,
		PriceCents: event.PriceCents,
		AddedAt:    time.Now(),
	}

	err := c.cartRepo.UpdateProductInCarts(ctx, &update)
	if err != nil {
		log.Printf("Failed to update carts for product %s: %v", event.ID, err)
		return err
	}

	log.Printf("Successfully updated product %s in all relevant carts", event.ID)
	return nil
}

func (c *CartConsumer) Close() error {
	log.Println("Close Kafka reader")
	return c.reader.Close()
}
