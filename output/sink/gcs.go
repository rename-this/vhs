package sink

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/gramLabs/vhs/session"
	"google.golang.org/api/option"
)

// GCSConfig is configuration for GCS
type GCSConfig struct {
	Session         *session.Session
	JSONKeyFilePath string
	ProjectID       string
	BucketName      string
}

// NewGCS creates a new Google Cloud Storage sink.
func NewGCS(ctx context.Context, cfg GCSConfig) (Sink, error) {
	c, err := storage.NewClient(ctx, option.WithCredentialsFile(cfg.JSONKeyFilePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	b := c.Bucket(cfg.BucketName)
	if err := b.Create(ctx, cfg.ProjectID, nil); err != nil {
		return nil, fmt.Errorf("failed to create bucket: %w", err)
	}

	return b.Object(cfg.Session.ID).NewWriter(ctx), nil
}
