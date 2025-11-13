package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CartItem struct {
	ProductID  string    `bson:"product_id" json:"product_id"`
	Name       string    `bson:"name" json:"name"`
	Image      string    `bson:"image" json:"image"`
	PriceCents int64     `bson:"price_cents" json:"price_cents"`
	Quantity   int64     `bson:"quantity" json:"quantity"`
	AddedAt    time.Time `bson:"added_at" json:"added_at"`
}

type Cart struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	UserID    string        `bson:"user_id" json:"user_id"`
	Items     []CartItem    `bson:"items,omitempty" json:"items,omitempty"`
	CreatedAt time.Time     `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt time.Time     `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}
