package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	model "notification-service/models"
	"notification-service/service"
	"time"

	"github.com/segmentio/kafka-go"
)

type NotificationConsumer struct {
	readers     []*kafka.Reader
	emailSender service.Notifier
}

func NewNotificationConsumer(brokers []string, topics []string, groupID string, emailSender service.Notifier) *NotificationConsumer {
	var readers []*kafka.Reader
	for _, topic := range topics {

		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		})
		readers = append(readers, reader)
	}
	return &NotificationConsumer{
		readers:     readers,
		emailSender: emailSender,
	}
}

func (n *NotificationConsumer) Consume(ctx context.Context) error {
	log.Println("Starting notification service")
	for _, r := range n.readers {
		go func(reader *kafka.Reader) {

			for {
				select {
				case <-ctx.Done():
					log.Println("Closing notification consumer for topic:", reader.Config().Topic)
					return
				default:
					msg, err := reader.FetchMessage(ctx)
					if err != nil {
						log.Println("fetch message err:", err)
						continue
					}

					log.Printf("Received message from topic [%s]\n", reader.Config().Topic)

					if err := n.ProcessMessage(ctx, msg); err != nil {
						log.Println("process message err:", err)
					} else {
						reader.CommitMessages(ctx, msg)
					}
				}
			}
		}(r)
	}
	<-ctx.Done()
	return nil
}
func (n *NotificationConsumer) ProcessMessage(ctx context.Context, msg kafka.Message) error {
	var eventType string
	for _, h := range msg.Headers {
		if h.Key == "event" {
			eventType = string(h.Value)
			break
		}
	}

	log.Println("PRocessing event... ", eventType)

	switch eventType {
	case "UserCreated":
		return n.handleUserCreated(ctx, msg.Value)
	case "OrderCreated":
		return n.handleOrderCreated(ctx, msg.Value)
	case "OrderShipped":
		return n.handleOrderShipped(ctx, msg.Value)
	case "PaymentCaptured":
		return n.handlePaymentCaptured(ctx, msg.Value)
	case "PaymentFailed":
		return n.handlePaymentFailed(ctx, msg.Value)
	default:
		log.Println("Ignoring event", eventType)
		return nil
	}
}

type OrderCreatedEvent struct {
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	UserEmail   string    `json:"user_email"`
	TotalAmount float64   `json:"total_amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	PlacedAt    time.Time `json:"placed_at"`
}

type OrderShippedEvent struct {
	OrderID   string `json:"order_id"`
	UserEmail string `json:"user_email"`
}

type PaymentCapturedEvent struct {
	PaymentID  string    `json:"payment_id"`
	OrderID    string    `json:"order_id"`
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	CapturedAt time.Time `json:"captured_at"`
	UserEmail  string    `json:"user_email"`
}

type PaymentFailedEvent struct {
	PaymentID string    `json:"payment_id"`
	OrderID   string    `json:"order_id"`
	Reason    string    `json:"reason"`
	FailedAt  time.Time `json:"failed_at"`
	UserEmail string    `json:"user_email"`
}

func (n *NotificationConsumer) handleUserCreated(ctx context.Context, data []byte) error {
	var event model.UserCreatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}
	log.Println("Sending user created confirmation mail", event.Name)
	emailReq := service.EmailRequest{
		To:      event.Email,
		Subject: fmt.Sprintf("Thank you for joining us - %s", event.Name),
		Body: fmt.Sprintf(`
		<div
  style="margin: 0 auto; font-family: Monospace; max-width: 600px;">
  <h2 style="color: #de64deff;">Welcome %s!</h2>
  <p>Thank you for joining us.</p>
  <p>We're hoping you will be shopping real soon. It's great to have you here!, btw</p>
</div>

		`, event.Name),
		Tags: []string{"user created", event.ID},
	}
	err := n.emailSender.SendEmail(ctx, emailReq)
	return err
}
func (n *NotificationConsumer) handleOrderCreated(ctx context.Context, data []byte) error {
	var event OrderCreatedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	log.Println("Sending order confirmation email", event.OrderID, event.UserEmail)

	emailReq := service.EmailRequest{
		To:      event.UserEmail,
		Subject: fmt.Sprintf("Order Confirmation - #%s", event.OrderID),
		Body: fmt.Sprintf(`
			<div style="font-family:Monospace; max-width: 600px; margin: 0 auto;">
				<h2 style="color: #de64deff;">Order Confirmed!</h2>
				<p>Got your order.</p>
				<div style="margin: 20px 0; background-color: #f5f5f535; padding: 15px; border-radius: 5px;">
					<p style="font-weight: bold;">Order ID: %s</p>
					<p style="font-weight: bold;">Total: $%.2f</p>
				</div>
				<p>Thank you!, btw</p>
			</div>
		`, event.OrderID, event.TotalAmount),
		Tags: []string{"order-confirmation", event.OrderID},
	}

	err := n.emailSender.SendEmail(ctx, emailReq)
	return err
}

func (n *NotificationConsumer) handleOrderShipped(ctx context.Context, data []byte) error {
	var event OrderShippedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}

	log.Println("Sending shipping notification", event.OrderID)

	emailReq := service.EmailRequest{
		To:      event.UserEmail,
		Subject: "Your Order Has Shipped!",
		Body: fmt.Sprintf(`
			<div style="font-family: Monospace; max-width: 600px; margin: 0 auto;">
				<h2 style="color: #de64deff;">Your Order is OWNING THE WAY IN</h2>
				<p>Your order has been shipped</p>
				<div style="background-color: #f5f5f535; padding: 15px; border-radius: 5px; margin: 20px 0;">
					<p style="font-weight: bold;">Order ID: %s</p>
				</div>
			</div>
		`, event.OrderID),
		Tags: []string{"order-shipped", event.OrderID},
	}

	err := n.emailSender.SendEmail(ctx, emailReq)
	return err
}

func (n *NotificationConsumer) handlePaymentCaptured(ctx context.Context, data []byte) error {
	var event PaymentCapturedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}
	emailReq := service.EmailRequest{
		To:      event.UserEmail,
		Subject: "Your payment is captured",
		Body: fmt.Sprintf(`
		<div style="font-family: Monospace; max-width: 600px; margin: 0 auto;">
				<h2 style="color: #de64deff;">Your Invoice will be with you soon!</h2>
				<p>Payment is Done!</p>
				<div style="background-color: #f5f5f535; padding: 15px; border-radius: 5px; margin: 20px 0;">
					<p style="font-weight: bold;">Payment ID: %s</p>
				</div>
			</div>
			payment captured
		`, event.PaymentID),
		Tags: []string("payment-captured", event.PaymentID),
	}
	log.Println("Payment captured", event.OrderID)
	err := n.emailSender.SendEmail(ctx, emailReq)
	return err
}

func (n *NotificationConsumer) handlePaymentFailed(ctx context.Context, data []byte) error {
	var event PaymentFailedEvent
	if err := json.Unmarshal(data, &event); err != nil {
		return err
	}
	emailReq := service.EmailRequest{
		To:      event.UserEmail,
		Subject: "Your payment is failed",
		Body: fmt.Sprintf(`
		<div style="font-family: Monospace; max-width: 600px; margin: 0 auto;">
				<h2 style="color: #de64deff;">Your payment didn't go through :/</h2>
				<p>Retry it again after some time to confirm you order...</p>
				<div style="background-color: #f5f5f535; padding: 15px; border-radius: 5px; margin: 20px 0;">
					<p style="font-weight: bold;">Payment ID: %s</p>
				</div>
			</div>
			payment failed
		`, event.PaymentID),
		Tags: []string("payment-failed", event.PaymentID),
	}
	log.Println("Payment failed", event.OrderID, event.Reason)
	err := n.emailSender.SendEmail(ctx, emailReq)
	return err
}

func (n *NotificationConsumer) Close() error {
	for _, r := range n.readers {
		if err := r.Close(); err != nil {
			return err
		}
	}
	return nil
}
