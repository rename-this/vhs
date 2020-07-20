package sink

import (
	"io/ioutil"
	"testing"

	"gotest.tools/v3/assert"
)

func TestHAR(t *testing.T) {
	cases := []struct {
		desc     string
		messages []interface{}
		out      harOut
	}{
		{
			desc: "no messages",
			out: harOut{
				Log: harLog{
					Version: "1.2",
					Creator: harCreator{
						Name:    "vhs",
						Version: "0.0.1",
					},
				},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			har := NewHAR(ioutil.Discard)
			for _, m := range c.messages {
				har.Write(m)
			}
			assert.DeepEqual(t, har.out, c.out)
		})
	}
}
