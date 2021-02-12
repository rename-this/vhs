package file

import (
	"fmt"
	"os"

	"github.com/rename-this/vhs/core"
)

// NewSource creates a new file source.
func NewSource(_ core.Context) (core.Source, error) {
	return &source{
		streams: make(chan core.InputReader),
	}, nil
}

type source struct {
	streams chan core.InputReader
}

func (s *source) Init(ctx core.Context) {
	ctx.Logger = ctx.Logger.With().
		Str(core.LoggerKeyComponent, "file_source").
		Logger()

	file, err := os.Open(ctx.FlowConfig.InputFile)
	if err != nil {
		ctx.Errors <- fmt.Errorf("failed to open %s: %w", ctx.FlowConfig.InputFile, err)
		return
	}

	s.streams <- &fileReader{
		file: file,
		meta: core.NewMeta(ctx.FlowConfig.InputFile, nil),
	}

	<-ctx.StdContext.Done()
}

func (s *source) Streams() <-chan core.InputReader {
	return s.streams
}

type fileReader struct {
	file *os.File
	meta *core.Meta
}

func (f *fileReader) Read(p []byte) (int, error) {
	return f.file.Read(p)
}

func (f *fileReader) Close() error {
	return f.file.Close()
}

func (f *fileReader) Meta() *core.Meta {
	return f.meta
}
