package plugin

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/rename-this/vhs/core"
	"github.com/rename-this/vhs/flow"
	"github.com/segmentio/ksuid"
	"gotest.tools/v3/assert"
)

func TestPlugin(t *testing.T) {
	cases := []struct {
		desc             string
		src              string
		applyErrContains string
		parser           func() *flow.Parser
		summary          Summary
	}{
		{
			desc:             "wrong sources type",
			src:              "package main\nfunc Sources() {}\n",
			applyErrContains: "failed to assert type",
			parser:           flow.NewParser,
		},
		{
			desc:   "new source",
			src:    sourcePlugin,
			parser: flow.NewParser,
		},
		{
			desc: "replace source",
			src:  sourcePlugin,
			parser: func() *flow.Parser {
				p := flow.NewParser()
				p.LoadSource("source", nil)
				return p
			},
			summary: Summary{Replaced: []string{"source"}},
		},
		{
			desc:             "wrong input modifier type",
			src:              "package main\nfunc InputModifiers() {}\n",
			applyErrContains: "failed to assert type",
			parser:           flow.NewParser,
		},
		{
			desc:   "new input modifier",
			src:    inputModifierPlugin,
			parser: flow.NewParser,
		},
		{
			desc: "replace input modifier",
			src:  inputModifierPlugin,
			parser: func() *flow.Parser {
				p := flow.NewParser()
				p.LoadInputModifier("input_modifier", nil)
				return p
			},
			summary: Summary{Replaced: []string{"input_modifier"}},
		},
		{
			desc:             "wrong input format type",
			src:              "package main\nfunc InputFormats() {}\n",
			applyErrContains: "failed to assert type",
			parser:           flow.NewParser,
		},
		{
			desc:   "new input format",
			src:    inputFormatPlugin,
			parser: flow.NewParser,
		},
		{
			desc: "replace input format",
			src:  inputFormatPlugin,
			parser: func() *flow.Parser {
				p := flow.NewParser()
				p.LoadInputFormat("input_format", nil)
				return p
			},
			summary: Summary{Replaced: []string{"input_format"}},
		},
		{
			desc:   "new output format",
			src:    outputFormatPlugin,
			parser: flow.NewParser,
		},
		{
			desc: "replace output format",
			src:  outputFormatPlugin,
			parser: func() *flow.Parser {
				p := flow.NewParser()
				p.LoadOutputFormat("output_format", nil)
				return p
			},
			summary: Summary{Replaced: []string{"output_format"}},
		},
		{
			desc:   "new output modifier",
			src:    outputModifierPlugin,
			parser: flow.NewParser,
		},
		{
			desc: "replace output modifier",
			src:  outputModifierPlugin,
			parser: func() *flow.Parser {
				p := flow.NewParser()
				p.LoadOutputModifier("output_modifier", nil)
				return p
			},
			summary: Summary{Replaced: []string{"output_modifier"}},
		},
		{
			desc:   "new sink",
			src:    sinkPlugin,
			parser: flow.NewParser,
		},
		{
			desc: "replace sink",
			src:  sinkPlugin,
			parser: func() *flow.Parser {
				p := flow.NewParser()
				p.LoadSink("sink", nil)
				return p
			},
			summary: Summary{Replaced: []string{"sink"}},
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			ctx := core.NewContext(&core.Config{Debug: true}, nil, nil)

			dir, err := ioutil.TempDir("", "")
			assert.NilError(t, err)
			defer os.RemoveAll(dir)

			err = ioutil.WriteFile(path.Join(dir, "main.go"), []byte(c.src), os.ModePerm)
			assert.NilError(t, err)

			wd, err := os.Getwd()
			assert.NilError(t, err)
			goModFile := fmt.Sprintf(goModFileFormat, ksuid.New().String(), path.Join(wd, ".."))

			err = ioutil.WriteFile(path.Join(dir, "go.mod"), []byte(goModFile), os.ModePerm)
			assert.NilError(t, err)

			cmd := exec.Command("go", "mod", "download", "-x")
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			if len(out) > 0 {
				fmt.Println(string(out))
			}
			assert.NilError(t, err)

			pluginFileName := fmt.Sprintf("plugin_%s.so", ksuid.New().String())

			// Build the plugin with the race detector so we can also
			// run the tests with it as well. Plugins must be built the same
			// exact way as the tests to be loaded.
			cmd = exec.Command("go", "build", "-race", "-buildmode=plugin", "-o", pluginFileName)
			cmd.Dir = dir
			out, err = cmd.CombinedOutput()
			if len(out) > 0 {
				fmt.Println(string(out))
			}
			assert.NilError(t, err)

			p, err := Load(path.Join(dir, pluginFileName))
			assert.NilError(t, err)

			parser := c.parser()
			summary, err := p.Apply(ctx, parser)
			if c.applyErrContains != "" {
				assert.ErrorContains(t, err, c.applyErrContains)
				return
			}
			assert.NilError(t, err)

			assert.DeepEqual(t, summary, c.summary)
		})
	}
}

const goModFileFormat = `
// Keep each plugin unique so we can load multiples.
module vhs_plugin_%s

go 1.15

require (
	github.com/rename-this/vhs v1.1.1
)

// This is a hack to make sure that we build the plugin
// with the local dev environment source.
replace github.com/rename-this/vhs v1.1.1 => %s
`

const sourcePlugin = `
package main

import (
	"github.com/rename-this/vhs/core"
)

func newTestSource(core.Context) (core.Source, error) { return &testSource{}, nil }
type testSource struct {}
func (*testSource) Streams() <-chan core.InputReader { return nil }
func (*testSource) Init(core.Context) {}

func Sources() map[string]core.SourceCtor { 
	return map[string]core.SourceCtor{
		"source": newTestSource,
	}
}
`

const inputModifierPlugin = `
package main

import (
	"github.com/rename-this/vhs/core"
)

func newTestInputModifier(core.Context) (core.InputModifier, error) { return &testInputModifier{}, nil }
type testInputModifier struct {}
func (*testInputModifier) Read([]byte) (int, error) { return -1, nil }
func (*testInputModifier) Wrap(core.InputReader) (core.InputReader, error) { return nil, nil }

func InputModifiers() map[string]core.InputModifierCtor { 
	return map[string]core.InputModifierCtor{
		"input_modifier": newTestInputModifier,
	}
}
`

const inputFormatPlugin = `
package main

import (
	"github.com/rename-this/vhs/core"
)

func newTestInputFormat(core.Context) (core.InputFormat, error) { return &testInputFormat{}, nil }
type testInputFormat struct {}
func (*testInputFormat) Init(core.Context, core.Middleware, <-chan core.InputReader) {}
func (*testInputFormat) Out() <-chan interface{} { return nil }

func InputFormats() map[string]core.InputFormatCtor { 
	return map[string]core.InputFormatCtor{
		"input_format": newTestInputFormat,
	}
}
`

const outputFormatPlugin = `
package main

import (
	"io"

	"github.com/rename-this/vhs/core"
)

func newTestOutputFormat(core.Context) (core.OutputFormat, error) { return &testOutputFormat{}, nil }
type testOutputFormat struct {}
func (*testOutputFormat) Init(core.Context, io.Writer) {}
func (*testOutputFormat) In() chan<- interface{} { return nil }
func (*testOutputFormat) Complete() <-chan struct{} { return nil }

func OutputFormats() map[string]core.OutputFormatCtor { 
	return map[string]core.OutputFormatCtor{
		"output_format": newTestOutputFormat,
	}
}
`

const outputModifierPlugin = `
package main

import (
	"github.com/rename-this/vhs/core"
)

func newTestOutputModifier(core.Context) (core.OutputModifier, error) { return &testOutputModifier{}, nil }
type testOutputModifier struct {}
func (*testOutputModifier) Wrap(core.OutputWriter) (core.OutputWriter, error) { return nil, nil }
func (*testOutputModifier) Write([]byte) (int, error) { return -1, nil }

func OutputModifiers() map[string]core.OutputModifierCtor { 
	return map[string]core.OutputModifierCtor{
		"output_modifier": newTestOutputModifier,
	}
}
`

const sinkPlugin = `
package main

import (
	"github.com/rename-this/vhs/core"
)

func newTestSink(core.Context) (core.Sink, error) { return &testSink{}, nil }
type testSink struct{}
func (*testSink) Write([]byte) (int, error) { return -1, nil }
func (*testSink) Close() error { return nil }


func Sinks() map[string]core.SinkCtor { 
	return map[string]core.SinkCtor{
		"sink": newTestSink,
	}
}
`
