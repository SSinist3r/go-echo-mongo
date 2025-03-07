// Package validator provides validation utilities for Echo applications.
package validator

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator is a custom validator for Echo
type CustomValidator struct {
	validator *validator.Validate
}

// New creates a new validator
func New() *CustomValidator {
	v := validator.New()

	// Register custom validation tags
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &CustomValidator{
		validator: v,
	}
}

// Validate validates the provided struct
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		// Convert validator errors to a map for better error messages
		validationErrors := err.(validator.ValidationErrors)
		errorMessages := make(map[string]string)

		for _, e := range validationErrors {
			errorMessages[e.Field()] = getErrorMessage(e)
		}

		return echo.NewHTTPError(http.StatusBadRequest, errorMessages)
	}
	return nil
}

// RegisterCustomValidation registers a custom validation function
func (cv *CustomValidator) RegisterCustomValidation(tag string, fn validator.Func) error {
	return cv.validator.RegisterValidation(tag, fn)
}

// getErrorMessage returns a human-readable error message for a validation error
func getErrorMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		if e.Type().Kind() == reflect.String {
			return "Must be at least " + e.Param() + " characters long"
		}
		return "Must be at least " + e.Param()
	case "max":
		if e.Type().Kind() == reflect.String {
			return "Must be at most " + e.Param() + " characters long"
		}
		return "Must be at most " + e.Param()
	case "oneof":
		return "Must be one of: " + e.Param()
	case "url":
		return "Invalid URL format"
	case "uuid":
		return "Invalid UUID format"
	case "alphanum":
		return "Must contain only alphanumeric characters"
	case "numeric":
		return "Must be a valid number"
	case "datetime":
		return "Invalid datetime format"
	}
	return "Invalid value"
}
