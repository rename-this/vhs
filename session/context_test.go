package session

import (
	"errors"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func TestContexts(t *testing.T) {
	var (
		canceled = make(chan struct{}, 3)
		errs     = make(chan error, 3)
		ctx1     = NewContexts(&Config{}, &FlowConfig{}, errs)
	)

	fn := func(ctx Context) {
		defer func() {
			ctx.Errors <- errors.New("111")
		}()
		for {
			select {
			case <-ctx.StdContext.Done():
				canceled <- struct{}{}
				return
			}
		}
	}

	go fn(ctx1)

	ctx1.Cancel()

	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, 1, len(canceled))
	assert.Equal(t, 1, len(errs))
}
