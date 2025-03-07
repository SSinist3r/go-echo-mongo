package dto

import (
	"time"

	"go-echo-mongo/internal/model"
)

// ProductResponse represents a product response
type ProductResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	Stock       int32     `json:"stock"`
	Category    string    `json:"category"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProductRequest represents the request body for creating a product
type CreateProductRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description string  `json:"description" validate:"required,min=10,max=1000"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int32   `json:"stock" validate:"required,gte=0"`
	Category    string  `json:"category" validate:"required"`
}

// UpdateProductRequest represents the request body for updating a product
type UpdateProductRequest struct {
	Name        string  `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description string  `json:"description,omitempty" validate:"omitempty,min=10,max=1000"`
	Price       float64 `json:"price,omitempty" validate:"omitempty,gt=0"`
	Stock       int32   `json:"stock,omitempty" validate:"omitempty,gte=0"`
	Category    string  `json:"category,omitempty" validate:"omitempty"`
}

// BatchCreateProductsRequest represents the request body for creating multiple products
type BatchCreateProductsRequest struct {
	Products []CreateProductRequest `json:"products" validate:"required,min=1,dive"`
}

// BatchUpdateProductsRequest represents the request body for updating multiple products
type BatchUpdateProductsRequest struct {
	Updates map[string]UpdateProductRequest `json:"updates" validate:"required,min=1"`
}

// BatchDeleteProductsRequest represents the request body for deleting multiple products
type BatchDeleteProductsRequest struct {
	IDs []string `json:"ids" validate:"required,min=1"`
}

// ProductFilterRequest represents the request body for filtering products
type ProductFilterRequest struct {
	Name     string  `json:"name,omitempty"`
	Category string  `json:"category,omitempty"`
	MinPrice float64 `json:"min_price,omitempty"`
	MaxPrice float64 `json:"max_price,omitempty"`
	Limit    int64   `json:"limit,omitempty"`
	Skip     int64   `json:"skip,omitempty"`
}

// ToModel converts CreateProductRequest to model.Product
func (r *CreateProductRequest) ToModel() *model.Product {
	return &model.Product{
		Name:        r.Name,
		Description: r.Description,
		Price:       r.Price,
		Stock:       r.Stock,
		Category:    r.Category,
	}
}

// ToModel converts UpdateProductRequest to model.Product
func (r *UpdateProductRequest) ToModel(existing *model.Product) *model.Product {
	if r.Name != "" {
		existing.Name = r.Name
	}
	if r.Description != "" {
		existing.Description = r.Description
	}
	if r.Price > 0 {
		existing.Price = r.Price
	}
	if r.Stock >= 0 {
		existing.Stock = r.Stock
	}
	if r.Category != "" {
		existing.Category = r.Category
	}
	return existing
}

// FromModel creates a ProductResponse from model.Product
func (r *ProductResponse) FromModel(product *model.Product) *ProductResponse {
	r.ID = product.ID.Hex()
	r.Name = product.Name
	r.Description = product.Description
	r.Price = product.Price
	r.Stock = product.Stock
	r.Category = product.Category
	r.CreatedAt = product.CreatedAt
	r.UpdatedAt = product.UpdatedAt
	return r
}

// NewProductResponse creates a new ProductResponse from model.Product
func NewProductResponse(product *model.Product) *ProductResponse {
	return new(ProductResponse).FromModel(product)
}

// NewProductResponseList creates a slice of ProductResponse from a slice of model.Product
func NewProductResponseList(products []*model.Product) []*ProductResponse {
	result := make([]*ProductResponse, len(products))
	for i, product := range products {
		result[i] = NewProductResponse(product)
	}
	return result
}

// ToModels converts BatchCreateProductsRequest to a slice of model.Product
func (r *BatchCreateProductsRequest) ToModels() []*model.Product {
	products := make([]*model.Product, len(r.Products))
	for i, productReq := range r.Products {
		products[i] = productReq.ToModel()
	}
	return products
}
