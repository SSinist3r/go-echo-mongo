package handler

import (
	"errors"
	"go-echo-mongo/internal/dto"
	"go-echo-mongo/internal/service"
	"go-echo-mongo/pkg/web/response"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// ProductHandler defines the interface for product-related HTTP handlers
type ProductHandler interface {
	Register(e *echo.Echo)
	Create(c echo.Context) error
	GetByID(c echo.Context) error
	GetAll(c echo.Context) error
	GetPaginated(c echo.Context) error
	GetByCategory(c echo.Context) error
	Update(c echo.Context) error
	Delete(c echo.Context) error

	// Batch operations
	CreateMany(c echo.Context) error
	FindByFilter(c echo.Context) error
	UpdateMany(c echo.Context) error
	DeleteMany(c echo.Context) error
}

// productHandler implements ProductHandler interface
type productHandler struct {
	service service.ProductService
}

// NewProductHandler creates a new ProductHandler instance
func NewProductHandler(service service.ProductService) ProductHandler {
	return &productHandler{
		service: service,
	}
}

// Register registers all product routes
func (h *productHandler) Register(e *echo.Echo) {
	products := e.Group("/api/v1/products")
	products.POST("", h.Create)
	products.GET("", h.GetAll)
	products.GET("/paginated", h.GetPaginated)
	products.GET("/:id", h.GetByID)
	products.PUT("/:id", h.Update)
	products.DELETE("/:id", h.Delete)
	products.GET("/category/:category", h.GetByCategory)

	// Batch operation routes
	products.POST("/batch", h.CreateMany)
	products.POST("/filter", h.FindByFilter)
	products.PUT("/batch", h.UpdateMany)
	products.DELETE("/batch", h.DeleteMany)
}

// Create handles product creation
func (h *productHandler) Create(c echo.Context) error {
	req := new(dto.CreateProductRequest)
	if err := c.Bind(req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return response.ValidationError(c, err)
	}

	product := req.ToModel()
	if err := h.service.Create(c.Request().Context(), product); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidStock):
			return response.BadRequest(c, "Stock cannot be negative")
		default:
			return response.InternalError(c, "Failed to create product")
		}
	}

	return response.Created(c, "Product created successfully", dto.NewProductResponse(product))
}

// GetByID handles retrieving a product by ID
func (h *productHandler) GetByID(c echo.Context) error {
	product, err := h.service.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			return response.NotFound(c, "Product not found")
		default:
			return response.InternalError(c, "Failed to retrieve product")
		}
	}

	return response.OK(c, "Product retrieved successfully", dto.NewProductResponse(product))
}

// GetAll handles retrieving all products
func (h *productHandler) GetAll(c echo.Context) error {
	products, err := h.service.GetAll(c.Request().Context())
	if err != nil {
		return response.InternalError(c, "Failed to retrieve products")
	}

	return response.OK(c, "Products retrieved successfully", dto.NewProductResponseList(products))
}

// GetByCategory handles retrieving products by category
func (h *productHandler) GetByCategory(c echo.Context) error {
	products, err := h.service.GetByCategory(c.Request().Context(), c.Param("category"))
	if err != nil {
		return response.InternalError(c, "Failed to retrieve products")
	}

	return response.OK(c, "Products retrieved successfully", dto.NewProductResponseList(products))
}

// Update handles updating a product
func (h *productHandler) Update(c echo.Context) error {
	req := new(dto.UpdateProductRequest)
	if err := c.Bind(req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return response.ValidationError(c, err)
	}
	// Get existing product first
	existingProduct, err := h.service.GetByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			return response.NotFound(c, "Product not found")
		default:
			return response.InternalError(c, "Failed to retrieve product")
		}
	}

	// Update only the fields that were provided
	updatedProduct := req.ToModel(existingProduct)
	if err := h.service.Update(c.Request().Context(), c.Param("id"), updatedProduct); err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			return response.NotFound(c, "Product not found")
		case errors.Is(err, service.ErrInvalidStock):
			return response.BadRequest(c, "Stock cannot be negative")
		default:
			return response.InternalError(c, "Failed to update product")
		}
	}

	return response.OK(c, "Product updated successfully", dto.NewProductResponse(updatedProduct))
}

// Delete handles deleting a product
func (h *productHandler) Delete(c echo.Context) error {
	if err := h.service.Delete(c.Request().Context(), c.Param("id")); err != nil {
		switch {
		case errors.Is(err, service.ErrProductNotFound):
			return response.NotFound(c, "Product not found")
		default:
			return response.InternalError(c, "Failed to delete product")
		}
	}

	return response.NoContent(c)
}

// CreateMany handles batch creation of products
func (h *productHandler) CreateMany(c echo.Context) error {
	req := new(dto.BatchCreateProductsRequest)
	if err := c.Bind(req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return response.ValidationError(c, err)
	}

	products := req.ToModels()
	if err := h.service.CreateProducts(c.Request().Context(), products); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidStock):
			return response.BadRequest(c, "One or more products have invalid stock values")
		case errors.Is(err, service.ErrEmptyBatch):
			return response.BadRequest(c, "No products provided")
		default:
			return response.InternalError(c, "Failed to create products")
		}
	}

	return response.Created(c, "Products created successfully", dto.NewProductResponseList(products))
}

// FindByFilter handles finding products by filter criteria
func (h *productHandler) FindByFilter(c echo.Context) error {
	req := new(dto.ProductFilterRequest)
	if err := c.Bind(req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}

	// Convert DTO to filter map
	filter := make(map[string]interface{})
	if req.Name != "" {
		filter["name"] = req.Name
	}
	if req.Category != "" {
		filter["category"] = req.Category
	}
	if req.MinPrice > 0 {
		filter["min_price"] = req.MinPrice
	}
	if req.MaxPrice > 0 {
		filter["max_price"] = req.MaxPrice
	}

	products, err := h.service.FindProductsByFilter(c.Request().Context(), filter, req.Limit, req.Skip)
	if err != nil {
		return response.InternalError(c, "Failed to find products")
	}

	return response.OK(c, "Products found successfully", dto.NewProductResponseList(products))
}

// UpdateMany handles batch update of products
func (h *productHandler) UpdateMany(c echo.Context) error {
	return response.NotImplemented(c, "Not implemented yet")
}

// DeleteMany handles batch deletion of products
func (h *productHandler) DeleteMany(c echo.Context) error {
	req := new(dto.BatchDeleteProductsRequest)
	if err := c.Bind(req); err != nil {
		return response.BadRequest(c, "Invalid request format")
	}

	if err := c.Validate(req); err != nil {
		return response.ValidationError(c, err)
	}

	count, err := h.service.DeleteProductsByIDs(c.Request().Context(), req.IDs)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmptyBatch):
			return response.BadRequest(c, "No valid product IDs provided")
		default:
			return response.InternalError(c, "Failed to delete products")
		}
	}

	return response.OK(c, "Products deleted successfully", map[string]int64{"deleted_count": count})
}

// GetPaginated handles the request to get products with pagination
func (h *productHandler) GetPaginated(c echo.Context) error {
	// Parse pagination parameters from query
	page, err := strconv.ParseInt(c.QueryParam("page"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	itemsPerPage, err := strconv.ParseInt(c.QueryParam("items_per_page"), 10, 64)
	if err != nil || itemsPerPage < 1 {
		itemsPerPage = 10
	}

	// Get products with pagination directly using the base service method
	products, totalCount, err := h.service.GetPaginated(
		c.Request().Context(),
		nil,
		page,
		itemsPerPage,
	)
	if err != nil {
		return response.InternalError(c, "Failed to retrieve products")
	}

	// Convert to response DTOs
	var responses []dto.ProductResponse
	for _, product := range products {
		responses = append(responses, dto.ProductResponse{
			ID:          product.ID.Hex(),
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Stock:       product.Stock,
			Category:    product.Category,
			CreatedAt:   product.CreatedAt,
			UpdatedAt:   product.UpdatedAt,
		})
	}

	// Return paginated response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": responses,
		"meta": map[string]interface{}{
			"current_page":   page,
			"items_per_page": itemsPerPage,
			"total_items":    totalCount,
			"total_pages":    (totalCount + itemsPerPage - 1) / itemsPerPage,
		},
	})
}
