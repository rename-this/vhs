package gzipx

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/gramLabs/vhs/internal/ioutilx"
	"gotest.tools/v3/assert"
)

var (
	decompressed = []byte("111")
	compressed   = []byte{31, 139, 8, 0, 0, 0, 0,
		0, 0, 255, 50, 52, 52, 4, 4, 0, 0, 255,
		255, 61, 81, 107, 77, 3, 0, 0, 0}
)

func TestNewOutputModifier(t *testing.T) {
	om, err := NewOutputModifier(nil)
	assert.NilError(t, err)

	var buf bytes.Buffer
	w, err := om.Wrap(ioutilx.NopWriteCloser(&buf))
	assert.NilError(t, err)

	_, err = w.Write(decompressed)
	assert.NilError(t, err)

	err = w.Close()
	assert.NilError(t, err)

	assert.DeepEqual(t, compressed, buf.Bytes())
}

func TestNewInputModifier(t *testing.T) {
	cases := []struct {
		desc        string
		in          []byte
		out         []byte
		errContains string
	}{
		{
			desc: "success",
			in:   compressed,
			out:  decompressed,
		},
		{
			desc:        "EOF",
			in:          []byte{},
			errContains: "EOF",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			im, err := NewInputModifier(nil)
			assert.NilError(t, err)

			buf := ioutilx.NopReadCloserID(ioutil.NopCloser(bytes.NewBuffer(c.in)))
			r, err := im.Wrap(buf)
			if c.errContains != "" {
				assert.ErrorContains(t, err, c.errContains)
				return
			}

			assert.NilError(t, err)

			assert.Equal(t, "", r.ID())

			b, err := ioutil.ReadAll(r)
			assert.NilError(t, err)

			err = r.Close()
			assert.NilError(t, err)

			assert.DeepEqual(t, c.out, b)
		})
	}
}
