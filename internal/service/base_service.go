package service

import (
	"context"
	"errors"
	"log"

	"go-echo-mongo/internal/model"
	"go-echo-mongo/internal/repository"

	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// Common service errors
	ErrNilContext    = errors.New("context cannot be nil")
	ErrNilRepository = errors.New("repository cannot be nil")
	ErrEmptyBatch    = errors.New("batch cannot be empty")

	// User service errors
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")

	// Product service errors
	ErrProductNotFound = errors.New("product not found")
	ErrInvalidStock    = errors.New("invalid stock value")
)

// BaseService provides common functionality for all services
type BaseService[T model.Model] interface {
	// Common CRUD operations
	Create(ctx context.Context, model T) error
	GetByID(ctx context.Context, id string) (T, error)
	GetAll(ctx context.Context) ([]T, error)
	GetPaginated(ctx context.Context, filter interface{}, page, itemsPerPage int64) ([]T, int64, error)
	Update(ctx context.Context, id string, model T) error
	Delete(ctx context.Context, id string) error

	// Batch operations
	CreateMany(ctx context.Context, models []T) error
	FindMany(ctx context.Context, filter interface{}, opts *options.FindOptions) ([]T, error)
	UpdateMany(ctx context.Context, filter interface{}, update interface{}) (int64, error)
	DeleteMany(ctx context.Context, filter interface{}) (int64, error)
}

// baseService implements common service functionality
type baseService[T model.Model] struct {
	repo repository.BaseRepository[T]
}

// newBaseService creates a new base service instance
func newBaseService[T model.Model](repo repository.BaseRepository[T]) BaseService[T] {
	if repo == nil {
		log.Fatal(ErrNilRepository)
	}
	return &baseService[T]{
		repo: repo,
	}
}

// validateContext checks if context is nil
func validateContext(ctx context.Context) error {
	if ctx == nil {
		return ErrNilContext
	}
	return nil
}

// Create implements generic create operation
func (s *baseService[T]) Create(ctx context.Context, model T) error {
	if err := validateContext(ctx); err != nil {
		return err
	}
	return s.repo.Create(ctx, model)
}

// GetByID implements generic get by ID operation
func (s *baseService[T]) GetByID(ctx context.Context, id string) (T, error) {
	var empty T
	if err := validateContext(ctx); err != nil {
		return empty, err
	}
	return s.repo.FindByID(ctx, id)
}

// GetAll implements generic get all operation
func (s *baseService[T]) GetAll(ctx context.Context) ([]T, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	return s.repo.FindAll(ctx)
}

// GetPaginated retrieves models with pagination
func (s *baseService[T]) GetPaginated(ctx context.Context, filter interface{}, page, itemsPerPage int64) ([]T, int64, error) {
	if err := validateContext(ctx); err != nil {
		return nil, 0, err
	}
	return s.repo.FindPaginated(ctx, filter, page, itemsPerPage)
}

// Update implements generic update operation
func (s *baseService[T]) Update(ctx context.Context, id string, model T) error {
	if err := validateContext(ctx); err != nil {
		return err
	}
	return s.repo.Update(ctx, id, model)
}

// Delete implements generic delete operation
func (s *baseService[T]) Delete(ctx context.Context, id string) error {
	if err := validateContext(ctx); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

// CreateMany implements batch create operation
func (s *baseService[T]) CreateMany(ctx context.Context, models []T) error {
	if err := validateContext(ctx); err != nil {
		return err
	}
	if len(models) == 0 {
		return ErrEmptyBatch
	}
	return s.repo.InsertMany(ctx, models)
}

// FindMany implements batch find operation
func (s *baseService[T]) FindMany(ctx context.Context, filter interface{}, opts *options.FindOptions) ([]T, error) {
	if err := validateContext(ctx); err != nil {
		return nil, err
	}
	return s.repo.FindMany(ctx, filter, opts)
}

// UpdateMany implements batch update operation
func (s *baseService[T]) UpdateMany(ctx context.Context, filter interface{}, update interface{}) (int64, error) {
	if err := validateContext(ctx); err != nil {
		return 0, err
	}
	return s.repo.UpdateMany(ctx, filter, update)
}

// DeleteMany implements batch delete operation
func (s *baseService[T]) DeleteMany(ctx context.Context, filter interface{}) (int64, error) {
	if err := validateContext(ctx); err != nil {
		return 0, err
	}
	return s.repo.DeleteMany(ctx, filter)
}
