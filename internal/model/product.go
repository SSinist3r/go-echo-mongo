package model

// Product represents the product model in the system
type Product struct {
	BaseModel   `bson:",inline"`
	Name        string  `json:"name" bson:"name" validate:"required,min=2,max=100"`
	Description string  `json:"description" bson:"description" validate:"required,min=10,max=1000"`
	Price       float64 `json:"price" bson:"price" validate:"required,gt=0"`
	Stock       int32   `json:"stock" bson:"stock" validate:"required,gte=0"`
	Category    string  `json:"category" bson:"category" validate:"required"`
}

// Ensure Product implements BaseModel interface
var _ Model = (*Product)(nil)
