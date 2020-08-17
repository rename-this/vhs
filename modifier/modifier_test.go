package modifier

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/gramLabs/vhs/ioutilx"
	"github.com/gramLabs/vhs/testhelper"
	"gotest.tools/v3/assert"
)

func TestWriteCloser(t *testing.T) {
	cases := []struct {
		desc         string
		writeClosers WriteClosers
		in           string
		out          string
	}{
		{
			desc: "no modifiers",
			in:   "111",
			out:  "111",
		},
		{
			desc: "single modifiers",
			in:   "1",
			writeClosers: WriteClosers{
				&testhelper.DoubleOutput{},
			},
			out: "11",
		},
		{
			desc: "multiple modifiers",
			in:   "1",
			writeClosers: WriteClosers{
				&testhelper.DoubleOutput{},
				&testhelper.DoubleOutput{},
				&testhelper.DoubleOutput{},
			},
			out: "11111111",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			var buf bytes.Buffer
			w, closers, err := c.writeClosers.Wrap(ioutilx.NopWriteCloser(&buf))
			assert.NilError(t, err)
			assert.Equal(t, len(c.writeClosers), len(closers))

			_, err = w.Write([]byte(c.in))
			assert.NilError(t, err)

			err = closers.Close()
			assert.NilError(t, err)

			assert.Equal(t, c.out, buf.String())
		})
	}
}

func TestReadClosers(t *testing.T) {
	cases := []struct {
		desc        string
		readClosers ReadClosers
		in          string
		out         string
	}{
		{
			desc: "no modifiers",
			in:   "111",
			out:  "111",
		},
		{
			desc: "single modifiers",
			in:   "1",
			readClosers: ReadClosers{
				&testhelper.DoubleInput{},
			},
			out: "11",
		},
		{
			desc: "multiple modifiers",
			in:   "1",
			readClosers: ReadClosers{
				&testhelper.DoubleInput{},
				&testhelper.DoubleInput{},
				&testhelper.DoubleInput{},
			},
			out: "11111111",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			buf := ioutil.NopCloser(bytes.NewBufferString(c.in))
			r, closers, err := c.readClosers.Wrap(buf)
			assert.NilError(t, err)
			assert.Equal(t, len(c.readClosers), len(closers))

			b, err := ioutil.ReadAll(r)
			assert.NilError(t, err)

			err = closers.Close()
			assert.NilError(t, err)

			assert.Equal(t, c.out, string(b))
		})
	}
}
