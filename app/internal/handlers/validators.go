package handlers

import (
	"encoding/base64"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// RegisterCustomValidators registers custom validators for Gin
func RegisterCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("base64", validateBase64)
		v.RegisterValidation("uuid", validateUUID)
	}
}

// validateBase64 checks if a string is valid base64
func validateBase64(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // empty values are checked separately through required
	}
	_, err := base64.StdEncoding.DecodeString(value)
	return err == nil
}

// validateUUID checks if a string is a valid UUID
func validateUUID(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // empty values are checked separately through required
	}
	_, err := uuid.Parse(value)
	return err == nil
}
