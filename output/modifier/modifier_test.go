package modifier

import (
	"bytes"
	"encoding/base64"
	"io"
	"testing"

	"gotest.tools/v3/assert"
)

type base64Mod struct{}

func (*base64Mod) Wrap(w io.WriteCloser) io.WriteCloser {
	return base64.NewEncoder(base64.URLEncoding, w)
}

type removeEMod struct{}

func (*removeEMod) Wrap(w io.WriteCloser) io.WriteCloser {
	return NopWriteCloser(&removeE{w: w})
}

type removeE struct {
	w io.WriteCloser
}

func (r *removeE) Write(p []byte) (int, error) {
	return r.w.Write(bytes.ReplaceAll(p, []byte{'E'}, []byte{}))
}

func TestModifiers(t *testing.T) {
	cases := []struct {
		desc string
		ms   Modifiers
		in   string
		out  string
	}{
		{
			desc: "no modifiers",
			in:   "111111",
			out:  "111111",
		},
		{
			desc: "single modifier",
			ms: Modifiers{
				&base64Mod{},
			},
			in:  "111111",
			out: "MTExMTEx",
		},
		{
			desc: "mutiple modifiers",
			ms: Modifiers{
				&base64Mod{},
				&removeEMod{},
			},
			in:  "111111",
			out: "MTxMTx",
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			var buf bytes.Buffer
			w, closeAll := c.ms.Wrap(NopWriteCloser(&buf))

			defer closeAll()

			w.Write([]byte(c.in))

			assert.DeepEqual(t, buf.String(), c.out)
		})
	}
}
