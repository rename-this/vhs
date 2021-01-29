package plugin

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"testing"

	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
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
			ctx := session.NewContexts(&session.Config{Debug: true}, nil, nil)

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
	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
)

func newTestSource(session.Context) (flow.Source, error) { return &testSource{}, nil }
type testSource struct {}
func (*testSource) Streams() <-chan flow.InputReader { return nil }
func (*testSource) Init(session.Context) {}

func Sources() map[string]flow.SourceCtor { 
	return map[string]flow.SourceCtor{
		"source": newTestSource,
	}
}
`

const inputModifierPlugin = `
package main

import (
	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
)

func newTestInputModifier(session.Context) (flow.InputModifier, error) { return &testInputModifier{}, nil }
type testInputModifier struct {}
func (*testInputModifier) Read([]byte) (int, error) { return -1, nil }
func (*testInputModifier) Wrap(flow.InputReader) (flow.InputReader, error) { return nil, nil }

func InputModifiers() map[string]flow.InputModifierCtor { 
	return map[string]flow.InputModifierCtor{
		"input_modifier": newTestInputModifier,
	}
}
`

const inputFormatPlugin = `
package main

import (
	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
	"github.com/rename-this/vhs/middleware"
)

func newTestInputFormat(session.Context) (flow.InputFormat, error) { return &testInputFormat{}, nil }
type testInputFormat struct {}
func (*testInputFormat) Init(session.Context, middleware.Middleware, <-chan flow.InputReader) {}
func (*testInputFormat) Out() <-chan interface{} { return nil }

func InputFormats() map[string]flow.InputFormatCtor { 
	return map[string]flow.InputFormatCtor{
		"input_format": newTestInputFormat,
	}
}
`

const outputFormatPlugin = `
package main

import (
	"io"

	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
)

func newTestOutputFormat(session.Context) (flow.OutputFormat, error) { return &testOutputFormat{}, nil }
type testOutputFormat struct {}
func (*testOutputFormat) Init(session.Context, io.Writer) {}
func (*testOutputFormat) In() chan<- interface{} { return nil }

func OutputFormats() map[string]flow.OutputFormatCtor { 
	return map[string]flow.OutputFormatCtor{
		"output_format": newTestOutputFormat,
	}
}
`

const outputModifierPlugin = `
package main

import (
	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
)

func newTestOutputModifier(session.Context) (flow.OutputModifier, error) { return &testOutputModifier{}, nil }
type testOutputModifier struct {}
func (*testOutputModifier) Wrap(flow.OutputWriter) (flow.OutputWriter, error) { return nil, nil }
func (*testOutputModifier) Write([]byte) (int, error) { return -1, nil }

func OutputModifiers() map[string]flow.OutputModifierCtor { 
	return map[string]flow.OutputModifierCtor{
		"output_modifier": newTestOutputModifier,
	}
}
`

const sinkPlugin = `
package main

import (
	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
)

func newTestSink(session.Context) (flow.Sink, error) { return &testSink{}, nil }
type testSink struct{}
func (*testSink) Write([]byte) (int, error) { return -1, nil }
func (*testSink) Close() error { return nil }


func Sinks() map[string]flow.SinkCtor { 
	return map[string]flow.SinkCtor{
		"sink": newTestSink,
	}
}
`
