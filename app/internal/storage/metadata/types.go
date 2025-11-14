package metadata

import "time"

// PayloadType captures the original form of the secret payload.
type PayloadType string

const (
	// PayloadTypeFile marks payloads that originate from file uploads.
	PayloadTypeFile PayloadType = "file"
	// PayloadTypeText marks payloads that originate from plain text inputs.
	PayloadTypeText PayloadType = "text"
)

// MetadataRecord describes metadata stored alongside ciphertext.
type MetadataRecord struct {
	ID           string      `json:"id"`
	FileName     string      `json:"fileName"`
	ContentType  string      `json:"contentType"`
	Size         int64       `json:"size"`
	ExpiresAt    time.Time   `json:"expiresAt"`
	PayloadType  PayloadType `json:"payloadType"`
	OpaqueRecord []byte      `json:"opaqueRecord,omitempty"`
}
