package driver

import (
	"context"
	"errors"
	"io"
	"time"
)

var ErrReferenceUnsupported = errors.New("storage driver does not support direct references")

type Reference struct {
	URL       string `json:"url"`
	ExpiresAt int64  `json:"expires_at"`
}

type Driver interface {
	Name() string
	Put(ctx context.Context, objectKey string, contentType string, body io.Reader) (int64, error)
	Open(ctx context.Context, objectKey string) (io.ReadCloser, error)
	Delete(ctx context.Context, objectKey string) error
	Reference(ctx context.Context, objectKey string, expiresAt time.Time) (*Reference, error)
}
