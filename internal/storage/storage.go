package storage

import (
	//"io"
	"time"
)

type Storage interface {
	// Generate presigned URL for upload (PUT)
	GenerateUploadURL(objectPath string, expires time.Duration) (string, error)

	// Generate presigned URL for download (GET)
	GenerateDownloadURL(objectPath string, expires time.Duration) (string, error)

	// Optional: used only by LocalStorage backend
	// SaveFile(objectPath string, data io.Reader) error
}
