package gcs

import (
	"cloud.google.com/go/storage"
	"github.com/go-errors/errors"
	"github.com/gramLabs/vhs/session"
	"github.com/gramLabs/vhs/sink"
	"github.com/gramLabs/vhs/source"
)

// NewSink creates a new Google Cloud Storage sink.
func NewSink(ctx *session.Context) (sink.Sink, error) {
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

// NewSource creates a new Google Cloud Storage source.
func NewSource(ctx *session.Context) (source.Source, error) {
	return nil, nil
}
