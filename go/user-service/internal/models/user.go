package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email     string             `bson:"email" json:"email" validate:"required,email"`
	Password  string             `bson:"password" json:"-" validate:"required,min=6"`
	Name      string             `bson:"name" json:"name" validate:"required,min=3"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
type UserCreatedEvent struct {
	ID    string    `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
	Time  time.Time `json:"time"`
}
