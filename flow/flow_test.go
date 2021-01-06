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
		InputDrainDuration: 50 * time.Millisecond,
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

		sink1 = &testSinkInt{}
		o1    = NewOutput(newTestOutputFormatNoErr(ctx), nil, sink1)

		mos = OutputModifiers{
			&TestDoubleOutputModifier{},
		}
		sink2 = &testSinkInt{}
		o2    = NewOutput(newTestOutputFormatNoErr(ctx), mos, sink2)

		oo = Outputs{o1, o2}

		f = &Flow{i, oo}
	)

	f.Run(ctx, nil)

	assert.Equal(t, 0, len(errs))

	sort.Ints(sink1.data)
	sort.Ints(sink2.data)

	assert.DeepEqual(t, sink1.data, []int{1, 1, 2, 2, 3, 3})
	assert.DeepEqual(t, sink2.data, []int{11, 11, 22, 22, 33, 33})
}
