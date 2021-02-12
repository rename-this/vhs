package flow

import (
	"fmt"
	"sync"
	"time"

	"github.com/rename-this/vhs/core"
	"github.com/rename-this/vhs/internal/observe"
)

// Input joins a source and a format with
// optional modifiers.
type Input struct {
	Source    core.Source
	Modifiers core.InputModifiers
	Format    core.InputFormat

	done chan struct{}
}

// NewInput creates a new input.
func NewInput(s core.Source, mods core.InputModifiers, f core.InputFormat) *Input {
	return &Input{
		Source:    s,
		Modifiers: mods,
		Format:    f,
		done:      make(chan struct{}),
	}
}

// Done returns a channel indicating that this input is done.
func (i *Input) Done() <-chan struct{} {
	return i.done
}

// Init starts the input.
func (i *Input) Init(ctx core.Context, m core.Middleware) {
	ctx.Logger = ctx.Logger.With().
		Str(core.LoggerKeyComponent, "input").
		Logger()

	ctx.Logger.Debug().Msg("init")

	// This is an interim channel that allows
	// the streams to be wrapped before passed
	// to the format.
	streams := make(chan core.InputReader)

	go i.Format.Init(ctx, m, streams)
	go i.Source.Init(ctx)

	var wg sync.WaitGroup

	for {
		select {
		case rs, more := <-i.Source.Streams():
			if !more {
				ctx.Logger.Debug().Msg("no more source streams")
				wg.Wait()
				ctx.Logger.Debug().Msg("all source streams EOF")
				time.Sleep(ctx.FlowConfig.InputDrainDuration)
				i.done <- struct{}{}
				return
			}

			wg.Add(1)

			rc := observe.NewReadCloser(rs)
			go func() {
				<-rc.EOF()
				ctx.Logger.Debug().Msg("source stream EOF")
				wg.Done()
			}()

			rs = &wrappedStream{
				rc:   rc,
				meta: rs.Meta(),
			}

			r, err := i.Modifiers.Wrap(rs)
			if err != nil {
				ctx.Errors <- fmt.Errorf("failed to wrap source stream: %w", err)
				continue
			}
			streams <- r
		case <-ctx.StdContext.Done():
			ctx.Logger.Debug().Msg("context canceled")
			return
		}
	}
}

type wrappedStream struct {
	rc   observe.ReadCloser
	meta *core.Meta
}

func (s *wrappedStream) Read(p []byte) (int, error) {
	return s.rc.Read(p)
}

func (s *wrappedStream) Close() error {
	return s.rc.Close()
}

func (s *wrappedStream) Meta() *core.Meta {
	return s.meta
}
