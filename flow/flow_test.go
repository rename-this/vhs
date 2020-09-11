package flow

import (
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/gramLabs/vhs/internal/ioutilx"
	"github.com/gramLabs/vhs/session"
	"gotest.tools/v3/assert"
)

func TestFlow(t *testing.T) {
	cfg := &session.Config{
		FlowDuration:       500 * time.Millisecond,
		InputDrainDuration: 500 * time.Millisecond,
		ShutdownDuration:   500 * time.Millisecond,
	}
	errs := make(chan error, 1)
	ctx, inputCtx, outputCtx := session.NewContexts(cfg, errs)

	var (
		s = &testSource{
			streams: make(chan ioutilx.ReadCloserID),
			data: []ioutilx.ReadCloserID{
				ioutilx.NopReadCloserID(ioutil.NopCloser(strings.NewReader("1\n2\n3\n"))),
			},
		}
		ifmt, _ = newTestInputFormat(inputCtx)
		mis     = InputModifiers{
			&TestDoubleInputModifier{},
		}
		i = NewInput(s, mis, ifmt)

		ofmt1 = &testSink{}
		o1    = NewOutput(newTestOutputFormatNoErr(outputCtx), nil, ofmt1)

		mos = OutputModifiers{
			&TestDoubleOutputModifier{},
		}
		ofmt2 = &testSink{}
		o2    = NewOutput(newTestOutputFormatNoErr(outputCtx), mos, ofmt2)

		oo = Outputs{o1, o2}

		f = &Flow{i, oo}
	)

	f.Run(ctx, inputCtx, outputCtx, nil)

	assert.Equal(t, 0, len(errs))
	assert.Equal(t, string(ofmt1.data), "123123")
	assert.Equal(t, string(ofmt2.data), "112233112233")
}
