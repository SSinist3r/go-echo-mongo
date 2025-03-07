package validator

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

// Test struct for benchmarking
type User struct {
	Name     string `json:"name" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"required,gte=0,lte=120"`
	Password string `json:"password" validate:"required,min=8"`
}

// Direct validation without reflection
func validateUserDirect(u *User) error {
	if u.Name == "" || len(u.Name) < 3 || len(u.Name) > 50 {
		return ValidationError{Field: "name", Message: "invalid name length"}
	}
	if !IsEmail(u.Email) {
		return ValidationError{Field: "email", Message: "invalid email"}
	}
	if u.Age < 0 || u.Age > 120 {
		return ValidationError{Field: "age", Message: "invalid age"}
	}
	if len(u.Password) < 8 {
		return ValidationError{Field: "password", Message: "password too short"}
	}
	return nil
}

// ValidationError for direct validation
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// IsEmail is a simple email validation function
func IsEmail(email string) bool {
	// Simple check for demonstration
	return len(email) > 3 && contains(email, "@") && contains(email, ".")
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark reflection-based validation
func BenchmarkValidateReflection(b *testing.B) {
	v := validator.New()
	user := &User{
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      30,
		Password: "password123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Struct(user)
	}
}

// Benchmark direct validation without reflection
func BenchmarkValidateDirect(b *testing.B) {
	user := &User{
		Name:     "John Doe",
		Email:    "john@example.com",
		Age:      30,
		Password: "password123",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateUserDirect(user)
	}
}

// Benchmark both valid and invalid cases
func BenchmarkValidateReflectionInvalid(b *testing.B) {
	v := validator.New()
	user := &User{
		Name:     "J",             // Invalid name (too short)
		Email:    "invalid-email", // Invalid email
		Age:      150,             // Invalid age
		Password: "123",           // Invalid password
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Struct(user)
	}
}

func BenchmarkValidateDirectInvalid(b *testing.B) {
	user := &User{
		Name:     "J",             // Invalid name (too short)
		Email:    "invalid-email", // Invalid email
		Age:      150,             // Invalid age
		Password: "123",           // Invalid password
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateUserDirect(user)
	}
}
