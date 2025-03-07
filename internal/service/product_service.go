package service

import (
	"context"
	"log"

	"go-echo-mongo/internal/model"
	"go-echo-mongo/internal/repository"
	"go-echo-mongo/internal/repository/redisrepo"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ProductService defines the interface for product-related business logic
type ProductService interface {
	BaseService[*model.Product]
	GetByCategory(ctx context.Context, category string) ([]*model.Product, error)
	UpdateStock(ctx context.Context, id string, quantity int32) error

	// Batch operations
	CreateProducts(ctx context.Context, products []*model.Product) error
	FindProductsByFilter(ctx context.Context, filter map[string]interface{}, limit, skip int64) ([]*model.Product, error)
	UpdateProductsByFilter(ctx context.Context, filter map[string]interface{}, updates map[string]interface{}) (int64, error)
	DeleteProductsByIDs(ctx context.Context, ids []string) (int64, error)
}

type productService struct {
	BaseService[*model.Product]
	repo  repository.ProductRepository
	redis redisrepo.Repository
}

// NewProductService creates a new ProductService instance
func NewProductService(repo repository.ProductRepository, redis redisrepo.Repository) ProductService {
	if repo == nil {
		log.Fatal(ErrNilRepository)
	}
	return &productService{
		BaseService: newBaseService(repo),
		repo:        repo,
		redis:       redis,
	}
}

// validateStock checks if the stock value is valid
func validateStock(stock int32) error {
	if stock < 0 {
		return ErrInvalidStock
	}
	return nil
}

// Create overrides base Create to add stock validation
func (s *productService) Create(ctx context.Context, product *model.Product) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	if err := validateStock(product.Stock); err != nil {
		return err
	}

	return s.BaseService.Create(ctx, product)
}

// GetByCategory retrieves products by category
func (s *productService) GetByCategory(ctx context.Context, category string) ([]*model.Product, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	return s.repo.FindByCategory(ctx, category)
}

// Update overrides base Update to add stock validation
func (s *productService) Update(ctx context.Context, id string, updates *model.Product) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	if err := validateStock(updates.Stock); err != nil {
		return err
	}

	if _, err := s.GetByID(ctx, id); err != nil {
		return ErrProductNotFound
	}

	return s.BaseService.Update(ctx, id, updates)
}

// UpdateStock updates a product's stock quantity
func (s *productService) UpdateStock(ctx context.Context, id string, quantity int32) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	if err := validateStock(quantity); err != nil {
		return err
	}

	product, err := s.GetByID(ctx, id)
	if err != nil {
		return ErrProductNotFound
	}

	product.Stock = quantity

	return s.BaseService.Update(ctx, id, product)
}

// CreateProducts creates multiple products with validation
func (s *productService) CreateProducts(ctx context.Context, products []*model.Product) error {
	if err := validateContext(ctx); err != nil {
		return err
	}

	if len(products) == 0 {
		return ErrEmptyBatch
	}

	// Validate stock for all products
	for _, product := range products {
		if err := validateStock(product.Stock); err != nil {
			return err
		}
	}

	return s.BaseService.CreateMany(ctx, products)
}

// FindProductsByFilter finds products by filter criteria
func (s *productService) FindProductsByFilter(ctx context.Context, filter map[string]interface{}, limit, skip int64) ([]*model.Product, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}

	// Convert map to BSON filter
	bsonFilter := bson.M{}

	// Handle special price range filters
	if minPrice, ok := filter["min_price"]; ok && minPrice.(float64) > 0 {
		bsonFilter["price"] = bson.M{"$gte": minPrice}
		delete(filter, "min_price")
	}

	if maxPrice, ok := filter["max_price"]; ok && maxPrice.(float64) > 0 {
		if priceFilter, exists := bsonFilter["price"]; exists {
			priceFilter.(bson.M)["$lte"] = maxPrice
		} else {
			bsonFilter["price"] = bson.M{"$lte": maxPrice}
		}
		delete(filter, "max_price")
	}

	// Add remaining filters
	for k, v := range filter {
		if v != "" && k != "limit" && k != "skip" {
			bsonFilter[k] = v
		}
	}

	// Set options
	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(limit)
	}
	if skip > 0 {
		opts.SetSkip(skip)
	}

	return s.BaseService.FindMany(ctx, bsonFilter, opts)
}

// UpdateProductsByFilter updates multiple products matching the filter
func (s *productService) UpdateProductsByFilter(ctx context.Context, filter map[string]interface{}, updates map[string]interface{}) (int64, error) {
	if err := validateContext(ctx); err != nil {
		return 0, err
	}

	// Validate stock if it's being updated
	if stock, ok := updates["stock"]; ok {
		stockValue, isInt := stock.(int32)
		if isInt && stockValue < 0 {
			return 0, ErrInvalidStock
		}
	}

	// Convert maps to BSON
	bsonFilter := bson.M{}
	for k, v := range filter {
		if v != "" {
			bsonFilter[k] = v
		}
	}

	bsonUpdate := bson.M{"$set": updates}

	return s.BaseService.UpdateMany(ctx, bsonFilter, bsonUpdate)
}

// DeleteProductsByIDs deletes multiple products by their IDs
func (s *productService) DeleteProductsByIDs(ctx context.Context, ids []string) (int64, error) {
	if err := validateContext(ctx); err != nil {
		return 0, err
	}

	if len(ids) == 0 {
		return 0, ErrEmptyBatch
	}

	// Convert string IDs to ObjectIDs
	objectIDs := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		objectID, err := model.StringToObjectID(id)
		if err != nil {
			continue // Skip invalid IDs
		}
		objectIDs = append(objectIDs, objectID)
	}

	if len(objectIDs) == 0 {
		return 0, ErrEmptyBatch
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}

	return s.BaseService.DeleteMany(ctx, filter)
}
