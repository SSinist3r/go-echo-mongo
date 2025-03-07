package repository

import (
	"context"
	"fmt"
	"go-echo-mongo/internal/model"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BaseRepository represents the base repository interface with generic operations
type BaseRepository[T model.Model] interface {
	// GetCollection returns the MongoDB collection
	GetCollection() *mongo.Collection

	// Single document operations
	Create(ctx context.Context, model T) (err error)
	FindByID(ctx context.Context, id string) (model T, err error)
	FindAll(ctx context.Context) (model []T, err error)
	FindPaginated(ctx context.Context, filter interface{}, page, itemsPerPage int64) (models []T, totalCount int64, err error)
	Update(ctx context.Context, id string, model T) (err error)
	Delete(ctx context.Context, id string) (err error)

	// Batch operations
	InsertMany(ctx context.Context, models []T) (err error)
	FindMany(ctx context.Context, filter interface{}, opts *options.FindOptions) (model []T, err error)
	UpdateMany(ctx context.Context, filter interface{}, update interface{}) (modifiedCount int64, err error)
	DeleteMany(ctx context.Context, filter interface{}) (deletedCount int64, err error)
}

// baseRepository implements BaseRepository for MongoDB
type baseRepository[T model.Model] struct {
	collection *mongo.Collection
}

// newBaseRepository creates a new MongoDB repository instance
func newBaseRepository[T model.Model](collection *mongo.Collection) *baseRepository[T] {
	if collection == nil {
		log.Fatal("collection cannot be nil")
	}
	return &baseRepository[T]{
		collection: collection,
	}
}

// GetCollection returns the MongoDB collection
func (r *baseRepository[T]) GetCollection() *mongo.Collection {
	return r.collection
}

// Create inserts a new model into the database
func (r *baseRepository[T]) Create(ctx context.Context, model T) error {
	// Set both timestamps to the same time
	now := time.Now().UTC()
	model.SetCreatedAt(now)
	model.SetUpdatedAt(now)

	result, err := r.collection.InsertOne(ctx, model)
	if err != nil {
		return fmt.Errorf("failed to create model: %w", err)
	}

	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return fmt.Errorf("invalid ID type returned from MongoDB")
	}
	model.SetID(id)

	return nil
}

// FindByID retrieves a model by its ID
func (r *baseRepository[T]) FindByID(ctx context.Context, id string) (T, error) {
	var model T
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return model, fmt.Errorf("invalid ID format: %w", err)
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return model, fmt.Errorf("model not found with ID %s", id)
		}
		return model, fmt.Errorf("failed to find model: %w", err)
	}
	return model, nil
}

// FindAll retrieves all models
func (r *baseRepository[T]) FindAll(ctx context.Context) ([]T, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to execute find query: %w", err)
	}
	defer cursor.Close(ctx)

	var models []T
	if err = cursor.All(ctx, &models); err != nil {
		return nil, fmt.Errorf("failed to decode models: %w", err)
	}

	return models, nil
}

// FindPaginated retrieves models with simple pagination
func (r *baseRepository[T]) FindPaginated(ctx context.Context, filter interface{}, page, itemsPerPage int64) ([]T, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if itemsPerPage < 1 {
		itemsPerPage = 10 // Default items per page
	}

	if filter == nil {
		filter = bson.M{}
	}

	// Calculate skip value
	skip := (page - 1) * itemsPerPage

	// Get total count
	totalCount, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Set up options for pagination
	findOptions := options.Find().
		SetSkip(skip).
		SetLimit(itemsPerPage)

	// Execute the query
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute find query: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode the results
	var models []T
	if err = cursor.All(ctx, &models); err != nil {
		return nil, 0, fmt.Errorf("failed to decode models: %w", err)
	}

	return models, totalCount, nil
}

// Update updates a model in the database
func (r *baseRepository[T]) Update(ctx context.Context, id string, model T) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	model.SetUpdatedAt(time.Now().UTC())
	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": objectID}, model)
	if err != nil {
		return fmt.Errorf("failed to update model: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("model not found with ID %s", id)
	}

	return nil
}

// Delete removes a model from the database
func (r *baseRepository[T]) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid ID format: %w", err)
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return fmt.Errorf("failed to delete model: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("model not found with ID %s", id)
	}

	return nil
}

// InsertMany creates multiple documents
func (r *baseRepository[T]) InsertMany(ctx context.Context, models []T) error {
	if len(models) == 0 {
		return nil
	}

	now := time.Now().UTC()
	documents := make([]interface{}, len(models))
	for i, model := range models {
		model.SetCreatedAt(now)
		model.SetUpdatedAt(now)
		documents[i] = model
	}

	result, err := r.collection.InsertMany(ctx, documents)
	if err != nil {
		return fmt.Errorf("failed to insert models: %w", err)
	}

	// Set the generated IDs back to the models
	for i, insertedID := range result.InsertedIDs {
		id, ok := insertedID.(primitive.ObjectID)
		if !ok {
			return fmt.Errorf("invalid ID type returned from MongoDB for model at index %d", i)
		}
		models[i].SetID(id)
	}

	return nil
}

// FindMany retrieves documents based on filter
func (r *baseRepository[T]) FindMany(ctx context.Context, filter interface{}, opts *options.FindOptions) ([]T, error) {
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to execute find query: %w", err)
	}
	defer cursor.Close(ctx)

	var models []T
	if err = cursor.All(ctx, &models); err != nil {
		return nil, fmt.Errorf("failed to decode models: %w", err)
	}

	return models, nil
}

// UpdateMany modifies multiple documents matching the filter
func (r *baseRepository[T]) UpdateMany(ctx context.Context, filter interface{}, update interface{}) (int64, error) {
	// For BulkWrite, we expect a slice of write models
	writeModels, ok := update.([]mongo.WriteModel)
	if !ok {
		return 0, fmt.Errorf("update parameter must be a slice of mongo.WriteModel")
	}

	if len(writeModels) == 0 {
		return 0, nil
	}

	// Execute bulk write operation
	result, err := r.collection.BulkWrite(ctx, writeModels)
	if err != nil {
		return 0, fmt.Errorf("failed to execute bulk write: %w", err)
	}

	return result.ModifiedCount, nil
}

// DeleteMany removes multiple documents matching the filter
func (r *baseRepository[T]) DeleteMany(ctx context.Context, filter interface{}) (int64, error) {
	result, err := r.collection.DeleteMany(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to delete models: %w", err)
	}

	return result.DeletedCount, nil
}
