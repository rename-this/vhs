package gcs

import (
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
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

	b := c.Bucket(ctx.FlowConfig.GCSBucketName)
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
		streams:   make(chan flow.InputReader),
		newClient: newClient,
	}
}

type gcsSource struct {
	streams   chan flow.InputReader
	newClient newClientFn
}

func (s *gcsSource) Streams() <-chan flow.InputReader {
	return s.streams
}

func (s *gcsSource) Init(ctx session.Context) {
	defer close(s.streams)

	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "gcs_source").
		Logger()

	c, err := s.newClient(ctx)
	if err != nil {
		ctx.Errors <- fmt.Errorf("failed to create client: %w", err)
		return
	}

	ctx.Logger.Debug().Msg("client created")

	b := c.Bucket(ctx.FlowConfig.GCSBucketName)
	if _, err := b.Attrs(ctx.StdContext); err != nil {
		ctx.Errors <- fmt.Errorf("failed to find bucket: %w", err)
		return
	}

	ctx.Logger.Debug().Msg("bucket found")

	o := b.Object(ctx.FlowConfig.GCSObjectName)
	r, err := o.NewReader(ctx.StdContext)
	if err != nil {
		ctx.Errors <- fmt.Errorf("failed to create object reader: %w", err)
		return
	}

	ctx.Logger.Debug().Msg("reader created")

	s.streams <- &gcsStream{
		Reader: r,
		meta:   flow.NewMeta(ctx.FlowConfig.GCSObjectName, nil),
	}
}

type gcsStream struct {
	*storage.Reader
	meta *flow.Meta
}

func (s *gcsStream) Meta() *flow.Meta {
	return s.meta
}
