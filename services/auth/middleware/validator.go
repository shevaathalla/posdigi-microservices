package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// FieldError represents a single field validation error
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// validationErrorBody is the JSON body returned on validation failure.
// Because echo.HTTPError.Message is not a string here, Echo's default error
// handler will use it as the root JSON response directly — no extra nesting.
type validationErrorBody struct {
	Message string       `json:"message"`
	Errors  []FieldError `json:"errors"`
}

// CustomValidator wraps go-playground/validator for Echo
type CustomValidator struct {
	validator *validator.Validate
}

// NewCustomValidator creates a new CustomValidator instance.
// It registers JSON tag names so field errors use snake_case names (e.g.
// "email") instead of Go struct field names (e.g. "Email").
func NewCustomValidator() *CustomValidator {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
	return &CustomValidator{validator: v}
}

// Validate implements echo.Validator
func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			fieldErrors := make([]FieldError, len(ve))
			for i, fe := range ve {
				fieldErrors[i] = FieldError{
					Field:   fe.Field(),
					Message: humanizeTag(fe),
				}
			}
			return echo.NewHTTPError(http.StatusBadRequest, validationErrorBody{
				Message: "Validation failed",
				Errors:  fieldErrors,
			})
		}
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return nil
}

// humanizeTag converts a validator.FieldError into a readable message
func humanizeTag(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, fe.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, fe.Param())
	case "oneof":
		opts := strings.ReplaceAll(fe.Param(), " ", ", ")
		return fmt.Sprintf("%s must be one of: %s", field, opts)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}
