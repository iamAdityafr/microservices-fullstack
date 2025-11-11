package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Product struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Category    string             `bson:"category" json:"category"`
	Image       string             `bson:"image" json:"image"`
	PriceCents  int64              `bson:"price" json:"price"`
	Description string             `bson:"description" json:"description"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// Kafka Events
type ProductUpdatedEvent struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Category   string    `json:"category"`
	Image      string    `json:"image"`
	PriceCents int64     `json:"price_cents"`
	UpdatedAt  time.Time `json:"updated_at"`
}
type ProductDeletedEvent struct {
	ID        string    `json:"id"`
	DeletedAt time.Time `json:"deleted_at"`
}
