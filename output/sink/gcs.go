package sink

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/gramLabs/vhs/session"
)

// GCSConfig is configuration for GCS
type GCSConfig struct {
	Session    *session.Session
	ProjectID  string
	BucketName string
}

// NewGCS creates a new Google Cloud Storage sink.
func NewGCS(cfg GCSConfig) (Sink, error) {
	ctx := context.Background()

	c, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	if _, err := c.Bucket(cfg.BucketName).Attrs(ctx); err != nil {
		return nil, fmt.Errorf("failed to find bucket: %w", err)
	}

	return b.Object(cfg.Session.ID).NewWriter(ctx), nil
}
