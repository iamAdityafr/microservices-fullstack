package model

import "time"

type UserCreatedEvent struct {
	ID    string    `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
	Time  time.Time `json:"time"`
}
type EmailStatus struct {
	Email     string    `json:"email"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error"`
}

type OrderPlacedEvent struct {
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
}
