package sink

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/go-errors/errors"
	"github.com/gramLabs/vhs/session"
)

// NewGCS creates a new Google Cloud Storage sink.
func NewGCS(ctx *session.Context) (Sink, error) {
	stdCtx := context.Background()

	c, err := storage.NewClient(stdCtx)
	if err != nil {
		return nil, errors.Errorf("failed to create client: %w", err)
	}

	b := c.Bucket(ctx.Config.GCSBucketName)
	if _, err := b.Attrs(stdCtx); err != nil {
		return nil, errors.Errorf("failed to find bucket: %w", err)
	}

	return b.Object(ctx.SessionID).NewWriter(stdCtx), nil
}
