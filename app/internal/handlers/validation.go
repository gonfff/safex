package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationError is a struct representing a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FormatValidationErrors formats validation errors
func FormatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   fieldError.Field(),
				Message: getValidationMessage(fieldError),
			})
		}
	}

	return errors
}

// getValidationMessage returns a clear error message
func getValidationMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", fieldError.Field())
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", fieldError.Field())
	case "base64":
		return fmt.Sprintf("%s must be valid base64", fieldError.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s", fieldError.Field(), fieldError.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", fieldError.Field(), fieldError.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", fieldError.Field(), fieldError.Param())
	default:
		return fmt.Sprintf("%s is invalid", fieldError.Field())
	}
}

// HandleValidationError handles validation errors
func (h *HTTPHandlers) HandleValidationError(c *gin.Context, err error) {
	validationErrors := FormatValidationErrors(err)

	// For JSON API, return JSON with errors
	if strings.Contains(c.GetHeader("Content-Type"), "application/json") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "validation failed",
			"errors": validationErrors,
		})
		return
	}

	// For forms, show the first error
	if len(validationErrors) > 0 {
		h.renderCreateResult(c, http.StatusBadRequest, fmt.Errorf("%s", validationErrors[0].Message))
		return
	}

	h.renderCreateResult(c, http.StatusBadRequest, fmt.Errorf("validation failed"))
}
