package gcs

import (
	"cloud.google.com/go/storage"
	"github.com/go-errors/errors"
	"github.com/gramLabs/vhs/flow"
	"github.com/gramLabs/vhs/internal/ioutilx"
	"github.com/gramLabs/vhs/session"
)

type newClientFn func(*session.Context) (*storage.Client, error)

func newClient(ctx *session.Context) (*storage.Client, error) {
	return storage.NewClient(ctx.StdContext)
}

// NewSink creates a new Google Cloud Storage sink.
func NewSink(ctx *session.Context) (flow.Sink, error) {
	return newSink(ctx, newClient)
}

func newSink(ctx *session.Context, newClient newClientFn) (flow.Sink, error) {
	c, err := newClient(ctx)
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
func NewSource(ctx *session.Context) (flow.Source, error) {
	return newSource(ctx, newClient), nil
}

func newSource(ctx *session.Context, newClient newClientFn) flow.Source {
	return &gcsSource{
		streams:   make(chan ioutilx.ReadCloserID),
		newClient: newClient,
	}
}

type gcsSource struct {
	streams   chan ioutilx.ReadCloserID
	newClient newClientFn
}

func (s *gcsSource) Streams() <-chan ioutilx.ReadCloserID { return s.streams }

func (s *gcsSource) Init(ctx *session.Context) {
	c, err := s.newClient(ctx)
	if err != nil {
		ctx.Errors <- errors.Errorf("failed to create client: %w", err)
		return
	}

	b := c.Bucket(ctx.Config.GCSBucketName)
	if _, err := b.Attrs(ctx.StdContext); err != nil {
		ctx.Errors <- errors.Errorf("failed to find bucket: %w", err)
		return
	}

	o := b.Object(ctx.Config.GCSObjectName)
	r, err := o.NewReader(ctx.StdContext)
	if err != nil {
		ctx.Errors <- errors.Errorf("failed to create object reader: %w", err)
		return
	}

	s.streams <- &gcsStream{
		Reader: r,
		id:     ctx.Config.GCSObjectName,
	}
}

type gcsStream struct {
	*storage.Reader
	id string
}

func (s *gcsStream) ID() string { return s.id }
