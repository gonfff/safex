package metadata

import "time"

// MetadataRecord describes metadata stored alongside ciphertext.
type MetadataRecord struct {
	ID          string    `json:"id"`
	FileName    string    `json:"fileName"`
	ContentType string    `json:"contentType"`
	Size        int64     `json:"size"`
	ExpiresAt   time.Time `json:"expiresAt"`
}
