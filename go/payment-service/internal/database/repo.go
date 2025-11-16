package database

import (
	"context"
	"errors"
	"payment-service/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type PaymentRepository interface {
	CreatePayment(ctx context.Context, p *models.Payment) error
	GetPaymentByOrderID(ctx context.Context, orderID string) (*models.Payment, error)
	UpdatePaymentStatus(ctx context.Context, orderID string, status models.PaymentStatus, providerRef, failReason string) error
}

type mongoPaymentRepo struct {
	col *mongo.Collection
}

func NewMongoPaymentRepo(client *mongo.Client, dbName string) PaymentRepository {
	return &mongoPaymentRepo{
		col: client.Database(dbName).Collection("payments"),
	}
}
func (m *mongoPaymentRepo) CreatePayment(ctx context.Context, p *models.Payment) error {
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()

	_, err := m.col.InsertOne(ctx, p)
	return err
}

func (m *mongoPaymentRepo) GetPaymentByOrderID(ctx context.Context, orderID string) (*models.Payment, error) {
	var payment models.Payment
	err := m.col.FindOne(ctx, bson.M{"order_id": orderID}).Decode(&payment)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &payment, nil
}

func (m *mongoPaymentRepo) UpdatePaymentStatus(ctx context.Context, orderID string, status models.PaymentStatus, providerRef, failReason string) error {
	update := bson.M{
		"$set": bson.M{
			"status":       status,
			"provider_ref": providerRef,
			"fail_reason":  failReason,
			"updated_at":   time.Now(),
		},
	}

	res, err := m.col.UpdateOne(ctx, bson.M{"order_id": orderID, "provider_ref": providerRef}, update, options.UpdateOne().SetUpsert(false))
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}
