package file

import (
	"fmt"
	"os"

	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
)

// NewSource creates a new file source.
func NewSource(_ session.Context) (flow.Source, error) {
	return &source{
		streams: make(chan flow.InputReader),
	}, nil
}

type source struct {
	streams chan flow.InputReader
}

func (s *source) Init(ctx session.Context) {
	ctx.Logger = ctx.Logger.With().
		Str(session.LoggerKeyComponent, "file_source").
		Logger()

	file, err := os.Open(ctx.FlowConfig.InputFile)
	if err != nil {
		ctx.Errors <- fmt.Errorf("failed to open %s: %w", ctx.FlowConfig.InputFile, err)
		return
	}

	s.streams <- &fileReader{
		file: file,
		meta: flow.NewMeta(ctx.FlowConfig.InputFile, nil),
	}

	<-ctx.StdContext.Done()
}

func (s *source) Streams() <-chan flow.InputReader {
	return s.streams
}

type fileReader struct {
	file *os.File
	meta *flow.Meta
}

func (f *fileReader) Read(p []byte) (int, error) {
	return f.file.Read(p)
}

func (f *fileReader) Close() error {
	return f.file.Close()
}

func (f *fileReader) Meta() *flow.Meta {
	return f.meta
}
