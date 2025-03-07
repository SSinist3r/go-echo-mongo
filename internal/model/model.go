package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Model interface defines the common fields that all models should have
type Model interface {
	GetID() primitive.ObjectID
	SetID(primitive.ObjectID)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetCreatedAt(time.Time)
	SetUpdatedAt(time.Time)
}

// model implements Model interface with common fields
type BaseModel struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// GetID returns the ID of the model
func (m *BaseModel) GetID() primitive.ObjectID {
	return m.ID
}

// SetID sets the ID of the model
func (m *BaseModel) SetID(id primitive.ObjectID) {
	m.ID = id
}

// GetCreatedAt returns the creation timestamp
func (m *BaseModel) GetCreatedAt() time.Time {
	return m.CreatedAt
}

// GetUpdatedAt returns the update timestamp
func (m *BaseModel) GetUpdatedAt() time.Time {
	return m.UpdatedAt
}

// SetCreatedAt sets the creation timestamp
func (m *BaseModel) SetCreatedAt(t time.Time) {
	m.CreatedAt = t
}

// SetUpdatedAt sets the update timestamp
func (m *BaseModel) SetUpdatedAt(t time.Time) {
	m.UpdatedAt = t
}

// StringToObjectID converts a string ID to a primitive.ObjectID
func StringToObjectID(id string) (primitive.ObjectID, error) {
	return primitive.ObjectIDFromHex(id)
}
