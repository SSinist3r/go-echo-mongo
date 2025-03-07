package repository

import (
	"context"

	"go-echo-mongo/internal/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// ProductRepository defines the interface for product-related database operations
type ProductRepository interface {
	BaseRepository[*model.Product]
	FindByCategory(context.Context, string) ([]*model.Product, error)
}

// productRepository implements ProductRepository interface
type productRepository struct {
	BaseRepository[*model.Product]
}

// NewProductRepository creates a new ProductRepository instance
func NewProductRepository(db *mongo.Database) ProductRepository {
	return &productRepository{
		BaseRepository: newBaseRepository[*model.Product](db.Collection("products")),
	}
}

// FindByCategory retrieves all products in a specific category
func (r *productRepository) FindByCategory(ctx context.Context, category string) ([]*model.Product, error) {
	cursor, err := r.GetCollection().Find(ctx, bson.M{"category": category})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*model.Product
	if err = cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	return products, nil
}
