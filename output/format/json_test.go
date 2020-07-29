package format

import (
	"io/ioutil"
	"testing"

	"gotest.tools/assert"
)

type vehicle struct {
	Wheels int
}

func TestJSONReader(t *testing.T) {
	cases := []struct {
		desc string
		n    interface{}
		b    []byte
	}{
		{
			desc: "simple",
			n:    &vehicle{Wheels: 2},
			b:    []byte(`{"Wheels":2}`),
		},
		{
			desc: "slice",
			n: []*vehicle{
				{Wheels: 2},
				{Wheels: 4},
			},
			b: []byte(`[{"Wheels":2},{"Wheels":4}]`),
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			r := NewJSONReader(c.n)
			b, err := ioutil.ReadAll(r)
			assert.NilError(t, err)
			assert.Equal(t, string(c.b), string(b))
		})
	}
}