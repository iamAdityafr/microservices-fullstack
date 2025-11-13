package database

import (
	"cart-service/internal/models"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type CartRepository interface {
	GetCart(ctx context.Context, userID string) (*models.Cart, error)
	AddToCart(ctx context.Context, userID string, productData *models.CartItem) (*models.CartItem, error)
	RemoveFromCart(ctx context.Context, userID string, productID string) error
	UpdateProductInCarts(ctx context.Context, item *models.CartItem) error
}

type mongoRepo struct {
	col *mongo.Collection
}

func NewMongoRepo(client *mongo.Client, dbName string) *mongoRepo {
	col := client.Database(dbName).Collection("cart")
	return &mongoRepo{col: col}
}

func (repo *mongoRepo) GetCart(ctx context.Context, userID string) (*models.Cart, error) {
	var cart models.Cart

	err := repo.col.FindOne(ctx, bson.M{"user_id": userID}).Decode(&cart)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &models.Cart{
				UserID:    userID,
				Items:     []models.CartItem{},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to fetch cart: %v", err)
	}

	if cart.Items == nil {
		cart.Items = []models.CartItem{}
	}

	return &cart, nil
}

func (repo *mongoRepo) UpdateProductInCarts(ctx context.Context, item *models.CartItem) error {
	filter := bson.M{"items.product_id": item.ProductID}
	update := bson.M{
		"$set": bson.M{
			"items.$[elem].name":        item.Name,
			"items.$[elem].image":       item.Image,
			"items.$[elem].price_cents": item.PriceCents,
			"updated_at":                time.Now(),
		},
	}
	opts := options.UpdateMany().SetArrayFilters([]interface{}{bson.M{"elem.product_id": item.ProductID}}).SetUpsert(false)
	_, err := repo.col.UpdateMany(ctx, filter, update, opts)
	return err
}

func (repo *mongoRepo) AddToCart(ctx context.Context, userID string, item *models.CartItem) (*models.CartItem, error) {
	filter := bson.M{"user_id": userID, "items.product_id": item.ProductID}
	update := bson.M{
		"$inc": bson.M{"items.$.quantity": item.Quantity},
		"$set": bson.M{"updated_at": time.Now()},
	}
	res, err := repo.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	if res.MatchedCount == 0 {
		_, err := repo.col.UpdateOne(ctx,
			bson.M{"user_id": userID},
			bson.M{
				"$push":        bson.M{"items": item},
				"$set":         bson.M{"updated_at": time.Now()},
				"$setOnInsert": bson.M{"created_at": time.Now()},
			},
			options.UpdateOne().SetUpsert(true),
		)
		if err != nil {
			return nil, err
		}
	}

	return item, nil
}

func (repo *mongoRepo) RemoveFromCart(ctx context.Context, userID string, productID string) error {
	update := bson.M{
		"$pull": bson.M{
			"items": bson.M{"product_id": productID},
		},
		"$set": bson.M{
			"updated_at": time.Now(),
		},
	}

	res, err := repo.col.UpdateOne(ctx, bson.M{"user_id": userID}, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}

	return nil
}
