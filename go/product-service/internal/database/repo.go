package database

import (
	"context"
	"errors"
	"fmt"
	"product-service/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ProductRepository interface {
	GetAllProducts(ctx context.Context) ([]*models.Product, error)
	CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error)
	GetProductById(ctx context.Context, id string) (*models.Product, error)
	UpdateProduct(ctx context.Context, id string, changes map[string]interface{}) (*models.Product, error)
	DeleteProduct(ctx context.Context, id string) error
	SearchProduct(ctx context.Context, filter bson.M, limit, skip int64) ([]*models.Product, error)
}

type mongoProductRepo struct {
	col *mongo.Collection
}

func NewMongoRepo(client *mongo.Client, dbName string) *mongoProductRepo {
	col := client.Database(dbName).Collection("product")
	return &mongoProductRepo{col: col}
}
func (repo *mongoProductRepo) GetAllProducts(ctx context.Context) ([]*models.Product, error) {
	var products []*models.Product

	res, err := repo.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer res.Close(ctx)

	for res.Next(ctx) {
		var product models.Product
		if err := res.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	if err := res.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (repo *mongoProductRepo) CreateProduct(ctx context.Context, product *models.Product) (*models.Product, error) {
	if product.ID.IsZero() {
		product.ID = primitive.NewObjectID()
	}

	now := time.Now()
	product.CreatedAt = now
	product.UpdatedAt = now

	_, err := repo.col.InsertOne(ctx, product)
	if err != nil {
		return nil, err
	}

	return product, nil
}
func (repo *mongoProductRepo) GetProductById(ctx context.Context, id string) (*models.Product, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id format: %v", err)
	}
	var product models.Product
	err = repo.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("can't find the product")
		}
	}
	return &product, nil
}
func (repo *mongoProductRepo) UpdateProduct(ctx context.Context, id string, changes map[string]interface{}) (*models.Product, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid product id")
	}
	changes["updated_at"] = time.Now()
	update := bson.M{"$set": changes}

	result, err := repo.col.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		return nil, err
	}
	if result.MatchedCount == 0 {
		return nil, errors.New("product not found")
	}

	var updatedProduct models.Product
	if err := repo.col.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedProduct); err != nil {
		return nil, err
	}

	return &updatedProduct, nil
}
func (repo *mongoProductRepo) DeleteProduct(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid product id")
	}
	res, err := repo.col.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return errors.New("product not found")
	}
	return nil
}

func (repo *mongoProductRepo) SearchProduct(ctx context.Context, filter bson.M, limit, skip int64) ([]*models.Product, error) {
	var products []*models.Product

	findOpts := options.Find()
	if limit > 0 {
		findOpts.SetLimit(limit)
	}
	if skip > 0 {
		findOpts.SetSkip(skip)
	}
	res, err := repo.col.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, err
	}
	defer res.Close(ctx)
	for res.Next(ctx) {
		var product models.Product
		if err := res.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}
	if err := res.Err(); err != nil {
		return nil, err
	}
	return products, err
}
