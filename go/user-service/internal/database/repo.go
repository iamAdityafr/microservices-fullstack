package database

import (
	"context"
	"errors"
	"log"
	"time"
	"user-service/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserById(ctx context.Context, id string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id string) error
}

type mongoRepo struct {
	col *mongo.Collection
}

func NewMongoRepo(client *mongo.Client, dbName string) *mongoRepo {
	col := client.Database(dbName).Collection("users_db")

	idxModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, _ = col.Indexes().CreateOne(context.Background(), idxModel)

	return &mongoRepo{col: col}
}

func (repo *mongoRepo) CreateUser(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := repo.col.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return errors.New("user already existing")
		}
		return err
	}
	return nil
}

func (repo *mongoRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := repo.col.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	log.Printf("DECODING to %T", &user)
	return &user, nil
}

func (repo *mongoRepo) GetUserById(ctx context.Context, id string) (*models.User, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user id format")
	}
	var user models.User
	err = repo.col.FindOne(ctx, bson.M{"_id": objectId}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (repo *mongoRepo) UpdateUser(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()

	update := bson.M{
		"$set": bson.M{
			"name":       user.Name,
			"email":      user.Email,
			"updated_at": user.UpdatedAt,
		},
	}

	res, err := repo.col.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (repo *mongoRepo) DeleteUser(ctx context.Context, id string) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user id format")
	}

	res, err := repo.col.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return errors.New("user not found")
	}

	return nil
}
