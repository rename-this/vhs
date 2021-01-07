package observe

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"gotest.tools/v3/assert"
)

func TestReadCloser(t *testing.T) {
	cases := []struct {
		desc        string
		rc          io.ReadCloser
		errContains string
	}{
		{
			desc: "EOF",
			rc:   ioutil.NopCloser(strings.NewReader("a\nb\nc\n")),
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			r := NewReadCloser(c.rc)

			_, err := ioutil.ReadAll(r)

			if c.errContains != "" {
				assert.ErrorContains(t, err, c.errContains)
			}

			assert.Equal(t, 1, len(r.EOF()))
			assert.NilError(t, r.Close())
		})
	}
}
