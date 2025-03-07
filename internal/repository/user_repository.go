package repository

import (
	"context"
	"log"
	"time"

	"go-echo-mongo/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository defines the interface for user-related database operations
type UserRepository interface {
	BaseRepository[*model.User]
	FindByEmail(context.Context, string) (*model.User, error)
	FindByApiKey(context.Context, string) (*model.User, error)
}

// userRepository implements UserRepository interface
type userRepository struct {
	BaseRepository[*model.User]
}

// createUserIndexes creates indexes for the user collection
func createUserIndexes(collection *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "email", Value: 1}},
			Options: &options.IndexOptions{
				Unique:     &[]bool{true}[0],
				Background: &[]bool{true}[0],
			},
		},
		{
			Keys: bson.D{{Key: "api_key", Value: 1}},
			Options: &options.IndexOptions{
				Unique:     &[]bool{true}[0],
				Background: &[]bool{true}[0],
			},
		},
		{
			Keys: bson.D{{Key: "roles", Value: 1}},
			Options: &options.IndexOptions{
				Background: &[]bool{true}[0],
			},
		},
	}

	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Fatal(err)
	}
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(db *mongo.Database) UserRepository {
	collection := db.Collection("users")

	// Create indexes for the user collection if they don't exist
	createUserIndexes(collection)

	return &userRepository{
		BaseRepository: newBaseRepository[*model.User](collection),
	}
}

// FindByEmail retrieves a user by their email
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user *model.User
	err := r.GetCollection().FindOne(ctx, bson.M{"email": email}).Decode(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByApiKey retrieves a user by their API key
func (r *userRepository) FindByApiKey(ctx context.Context, apiKey string) (*model.User, error) {
	user := &model.User{}
	err := r.GetCollection().FindOne(ctx, bson.M{"api_key": apiKey}).Decode(user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return user, nil
}
