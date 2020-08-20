package sink

import (
	"cloud.google.com/go/storage"
	"github.com/go-errors/errors"
	"github.com/gramLabs/vhs/session"
)

// NewGCS creates a new Google Cloud Storage sink.
func NewGCS(ctx *session.Context) (Sink, error) {
	c, err := storage.NewClient(ctx.StdContext)
	if err != nil {
		return nil, errors.Errorf("failed to create client: %w", err)
	}

	b := c.Bucket(ctx.Config.GCSBucketName)
	if _, err := b.Attrs(ctx.StdContext); err != nil {
		return nil, errors.Errorf("failed to find bucket: %w", err)
	}

	return b.Object(ctx.SessionID).NewWriter(ctx.StdContext), nil
}
