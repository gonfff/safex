package handlers

import (
	"encoding/base64"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// RegisterCustomValidators регистрирует кастомные валидаторы
func RegisterCustomValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("base64", validateBase64)
		v.RegisterValidation("uuid", validateUUID)
	}
}

// validateBase64 проверяет валидность base64 строки
func validateBase64(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // пустые значения проверяются отдельно через required  
	}
	_, err := base64.StdEncoding.DecodeString(value)
	return err == nil
}

// validateUUID проверяет валидность UUID
func validateUUID(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	if value == "" {
		return true // пустые значения проверяются отдельно через required
	}
	_, err := uuid.Parse(value)
	return err == nil
}