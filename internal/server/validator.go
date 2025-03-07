package server

import (
	"go-echo-mongo/pkg/web/validator"

	"github.com/labstack/echo/v4"
)

// // Validator represents the request validator
// type Validator struct {
// 	validator *validator.Validate
// }

// // Validate implements echo.Validator interface
// func (v *Validator) Validate(i interface{}) error {
// 	return v.validator.Struct(i)
// }

// setupValidator configures the validator for the server
func setupValidator(e *echo.Echo) {
	v := validator.New()
	// v.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
	// 	return validation.IsStrongPassword(fl.Field().String())
	// })
	e.Validator = v
	// e.Validator = &Validator{validator: v}
}
