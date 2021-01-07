package file

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/rename-this/vhs/session"
	"gotest.tools/v3/assert"
)

func TestNewSource(t *testing.T) {
	cases := []struct {
		desc        string
		file        string
		errContains string
	}{
		{
			desc:        "no file",
			file:        "/no/such/file",
			errContains: "no such file or directory",
		},
		{
			desc: "read file",
			file: "../testdata/test.json",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			var (
				errs = make(chan error, 1)
				ctx  = session.NewContexts(&session.Config{}, &session.FlowConfig{InputFile: c.file}, errs)
			)

			s, err := NewSource(ctx)
			assert.NilError(t, err)

			go s.Init(ctx)

			time.Sleep(500 * time.Millisecond)

			ctx.Cancel()

			if c.errContains != "" {
				assert.Equal(t, 1, len(errs))
				assert.ErrorContains(t, <-errs, c.errContains)
				return
			}

			r := <-s.Streams()
			defer r.Close()

			assert.Equal(t, r.Meta().SourceID, c.file)

			_, err = ioutil.ReadAll(r)
			assert.NilError(t, err)
		})
	}
}
