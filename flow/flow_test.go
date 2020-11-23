package flow

import (
	"io/ioutil"
	"sort"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/rename-this/vhs/session"
)

func TestFlow(t *testing.T) {
	cfg := &session.Config{Debug: true}
	flowCfg := &session.FlowConfig{
		SourceDuration: 500 * time.Millisecond,
		DrainDuration:  500 * time.Millisecond,
	}
	errs := make(chan error, 1)
	ctx := session.NewContexts(cfg, flowCfg, errs)

	var (
		s = &testSource{
			streams: make(chan InputReader),
			data: []InputReader{
				EmptyMeta(ioutil.NopCloser(strings.NewReader("1\n2\n3\n"))),
			},
		}
		ifmt, _ = newTestInputFormat(ctx)
		mis     = InputModifiers{
			&TestDoubleInputModifier{},
		}
		i = NewInput(s, mis, ifmt)

		ofmt1 = &testSinkInt{}
		o1    = NewOutput(newTestOutputFormatNoErr(ctx), nil, ofmt1)

		mos = OutputModifiers{
			&TestDoubleOutputModifier{},
		}
		ofmt2 = &testSinkInt{}
		o2    = NewOutput(newTestOutputFormatNoErr(ctx), mos, ofmt2)

		oo = Outputs{o1, o2}

		f = &Flow{i, oo}
	)

	f.Run(ctx, nil)

	ctx.Cancel()

	time.Sleep(500 * time.Millisecond)

	assert.Equal(t, 0, len(errs))

	sort.Ints(ofmt1.data)
	sort.Ints(ofmt2.data)

	assert.DeepEqual(t, ofmt1.data, []int{1, 1, 2, 2, 3, 3})
	assert.DeepEqual(t, ofmt2.data, []int{11, 11, 22, 22, 33, 33})
}
