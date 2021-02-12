package flow

import (
	"io/ioutil"
	"sort"
	"strings"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"github.com/rename-this/vhs/core"
	"github.com/rename-this/vhs/coretest"
)

func TestFlow(t *testing.T) {
	cfg := &core.Config{Debug: true}
	flowCfg := &core.FlowConfig{
		InputDrainDuration: 50 * time.Millisecond,
	}

	errs := make(chan error, 1)
	ctx := core.NewContext(cfg, flowCfg, errs)

	var (
		s = coretest.NewTestSourceData([]core.InputReader{
			core.EmptyMeta(ioutil.NopCloser(strings.NewReader("1\n2\n3\n"))),
		})
		ifmt, _ = coretest.NewTestInputFormat(ctx)
		imods   = core.InputModifiers{
			&coretest.TestDoubleInputModifier{},
		}
		i = NewInput(s, imods, ifmt)

		sink1 = &coretest.TestSinkInt{}
		o1    = NewOutput(coretest.NewTestOutputFormatNoErr(ctx), nil, sink1)

		omods = core.OutputModifiers{
			&coretest.TestDoubleOutputModifier{},
		}
		sink2 = &coretest.TestSinkInt{}
		o2    = NewOutput(coretest.NewTestOutputFormatNoErr(ctx), omods, sink2)

		oo = Outputs{o1, o2}

		f = &Flow{i, oo}
	)

	f.Run(ctx, nil)

	assert.Equal(t, 0, len(errs))

	var (
		sink1Data = sink1.Data()
		sink2Data = sink2.Data()
	)
	sort.Ints(sink1Data)
	sort.Ints(sink2Data)

	assert.DeepEqual(t, sink1Data, []int{1, 1, 2, 2, 3, 3})
	assert.DeepEqual(t, sink2Data, []int{11, 11, 22, 22, 33, 33})
}
