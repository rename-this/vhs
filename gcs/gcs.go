package gcs

import (
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/gramLabs/vhs/flow"
	"github.com/gramLabs/vhs/internal/ioutilx"
	"github.com/gramLabs/vhs/session"
)

type newClientFn func(session.Context) (*storage.Client, error)

func newClient(ctx session.Context) (*storage.Client, error) {
	return storage.NewClient(ctx.StdContext)
}

// NewSink creates a new Google Cloud Storage sink.
func NewSink(ctx session.Context) (flow.Sink, error) {
	return newSink(ctx, newClient)
}

func newSink(ctx session.Context, newClient newClientFn) (flow.Sink, error) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "gcs_sink").
		Logger()

	ctx.Logger.Debug().Msg("creating sink")

	c, err := newClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	ctx.Logger.Debug().Msg("client created")

	b := c.Bucket(ctx.Config.GCSBucketName)
	if _, err := b.Attrs(ctx.StdContext); err != nil {
		return nil, fmt.Errorf("failed to find bucket: %w", err)
	}

	ctx.Logger.Debug().Msg("creating writer")

	return b.Object(ctx.SessionID).NewWriter(ctx.StdContext), nil
}

// NewSource creates a new Google Cloud Storage source.
func NewSource(ctx session.Context) (flow.Source, error) {
	return newSource(ctx, newClient), nil
}

func newSource(ctx session.Context, newClient newClientFn) flow.Source {
	ctx.Logger.Debug().Msg("creating gcs source")
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

func (s *gcsSource) Init(ctx session.Context) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "gcs_source").
		Logger()

	c, err := s.newClient(ctx)
	if err != nil {
		ctx.Errors <- fmt.Errorf("failed to create client: %w", err)
		return
	}

	ctx.Logger.Debug().Msg("client created")

	b := c.Bucket(ctx.Config.GCSBucketName)
	if _, err := b.Attrs(ctx.StdContext); err != nil {
		ctx.Errors <- fmt.Errorf("failed to find bucket: %w", err)
		return
	}

	ctx.Logger.Debug().Msg("bucket found")

	o := b.Object(ctx.Config.GCSObjectName)
	r, err := o.NewReader(ctx.StdContext)
	if err != nil {
		ctx.Errors <- fmt.Errorf("failed to create object reader: %w", err)
		return
	}

	ctx.Logger.Debug().Msg("reader created")

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
