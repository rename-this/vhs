package s3compat

import (
	"bytes"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rename-this/vhs/core"
)

// NewSink creates a new S3-compatible sink.
func NewSink(ctx core.Context) (core.Sink, error) {
	ctx.Logger = ctx.Logger.With().
		Str(core.LoggerKeyComponent, "s3compat_sink").
		Logger()

	ctx.Logger.Debug().Msg("creating sink")

	client, err := newClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("falied to create S3-compatible sink client: %w", err)
	}

	ctx.Logger.Debug().Msg("creating sink")

	return &sink{
		ctx:    ctx,
		client: client,
		buf:    &bytes.Buffer{},
	}, nil
}

type sink struct {
	ctx    core.Context
	client *minio.Client
	buf    *bytes.Buffer
}

func (s *sink) Write(p []byte) (int, error) {
	return s.buf.Write(p)
}

func (s *sink) Close() error {
	_, err := s.client.PutObject(
		s.ctx.StdContext,
		s.ctx.FlowConfig.S3CompatBucketName,
		s.ctx.SessionID,
		s.buf,
		int64(s.buf.Len()),
		minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to put object to S3-compatible store: %w", err)
	}

	s.ctx.Logger.Debug().Msg("sink closed")

	return nil
}

// NewSource creates a new S3-compatible source.
func NewSource(ctx core.Context) (core.Source, error) {
	client, err := newClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("falied to create S3-compatible source client: %w", err)
	}
	return &source{
		client:  client,
		streams: make(chan core.InputReader),
	}, nil
}

type source struct {
	client  *minio.Client
	streams chan core.InputReader
}

func (s *source) Init(ctx core.Context) {
	ctx.Logger = ctx.Logger.With().
		Str(core.LoggerKeyComponent, "s3compat_source").
		Logger()

	o, err := s.client.GetObject(
		ctx.StdContext,
		ctx.FlowConfig.S3CompatBucketName,
		ctx.FlowConfig.S3CompatObjectName,
		minio.GetObjectOptions{})
	if err != nil {
		ctx.Errors <- fmt.Errorf("failed to get object from S3-compatible store: %w", err)
		return
	}

	s.streams <- &stream{
		Object: o,
		meta:   core.NewMeta(ctx.FlowConfig.S3CompatObjectName, nil),
	}

	ctx.Logger.Debug().Msg("init complete")
}

func (s *source) Streams() <-chan core.InputReader {
	return s.streams
}

type stream struct {
	*minio.Object
	meta *core.Meta
}

func (s *stream) Meta() *core.Meta {
	return s.meta
}

func newClient(ctx core.Context) (*minio.Client, error) {
	return minio.New(ctx.FlowConfig.S3CompatEndpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			ctx.FlowConfig.S3CompatAccessKey,
			ctx.FlowConfig.S3CompatSecretKey,
			ctx.FlowConfig.S3CompatToken,
		),
		Secure: ctx.FlowConfig.S3CompatSecure,
	})
}
