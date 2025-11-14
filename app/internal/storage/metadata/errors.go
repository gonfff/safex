package metadata

import "errors"

var (
	ErrNotFound = errors.New("secret not found")
	ErrExpired  = errors.New("secret expired")
)
