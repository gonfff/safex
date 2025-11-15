package domain

import (
	"time"
)

// PayloadType describes the type of secret content
type PayloadType string

const (
	PayloadTypeFile PayloadType = "file"
	PayloadTypeText PayloadType = "text"
)

// Secret represents a secret in the domain model
type Secret struct {
	ID           string
	FileName     string
	ContentType  string
	Size         int64
	ExpiresAt    time.Time
	PayloadType  PayloadType
	OpaqueRecord []byte
	Payload      []byte
}

// IsExpired checks if the secret has expired
func (s *Secret) IsExpired() bool {
	return s.ExpiresAt.Before(time.Now())
}

// IsText checks if the secret is text-based
func (s *Secret) IsText() bool {
	return s.PayloadType == PayloadTypeText
}
