package flow

import (
	"encoding/json"
	"testing"

	"github.com/rename-this/vhs/core"
	"github.com/rename-this/vhs/coretest"
	"gotest.tools/v3/assert"
)

func newTestParser() *Parser {
	p := NewParser()

	p.LoadSource("src", coretest.NewTestSource)

	p.LoadInputModifier("dbl", func(core.Context) (core.InputModifier, error) {
		return &coretest.TestDoubleInputModifier{}, nil
	})

	p.LoadInputFormat("ifmt", coretest.NewTestInputFormat)

	p.LoadOutputFormat("ofmt", coretest.NewTestOutputFormat)

	p.LoadOutputModifier("dbl", func(_ core.Context) (core.OutputModifier, error) {
		return &coretest.TestDoubleOutputModifier{}, nil
	})

	p.LoadSink("snk", func(_ core.Context) (core.Sink, error) {
		return &coretest.TestSink{}, nil
	})

	return p
}

func TestParse(t *testing.T) {
	cases := []struct {
		desc        string
		inputLine   string
		outputLines []string
		errContains string
		// A hack, but easier for comparing stuff that
		// doesn't really have a representation.
		flowJSON string
	}{
		{
			desc:      "many modifiers",
			inputLine: "src|dbl|dbl|dbl|dbl|ifmt",
			outputLines: []string{
				"ofmt|dbl|dbl|dbl|dbl|snk",
				"ofmt|dbl|dbl|dbl|dbl|snk",
			},
			flowJSON: `{"Input":{"Source":{"Data":null},"Modifiers":[{"OptCloseErr":null},{"OptCloseErr":null},{"OptCloseErr":null},{"OptCloseErr":null}],"Format":{}},"Outputs":[{"Format":{},"Modifiers":[{"OptCloseErr":null},{"OptCloseErr":null},{"OptCloseErr":null},{"OptCloseErr":null}],"Sink":{"OptCloseErr":null}},{"Format":{},"Modifiers":[{"OptCloseErr":null},{"OptCloseErr":null},{"OptCloseErr":null},{"OptCloseErr":null}],"Sink":{"OptCloseErr":null}}]}`,
		},
		{
			desc:        "bad input",
			inputLine:   "---",
			errContains: "invalid source: ---",
		},
		{
			desc:      "bad output",
			inputLine: "src|dbl|ifmt",
			outputLines: []string{
				"---",
			},
			errContains: "invalid output format: ---",
		},
	}
	for _, c := range cases {
		parser := newTestParser()
		ctx := core.NewContext(&core.Config{}, &core.FlowConfig{}, nil)
		t.Run(c.desc, func(t *testing.T) {
			i, err := parser.Parse(ctx, c.inputLine, c.outputLines)
			if c.errContains == "" {
				assert.NilError(t, err)
				b, err := json.Marshal(i)
				assert.NilError(t, err)
				assert.Equal(t, c.flowJSON, string(b))
			} else {
				assert.ErrorContains(t, err, c.errContains)
			}
		})
	}
}

func TestParseInput(t *testing.T) {
	cases := []struct {
		desc string
		line string
		// A hack, but easier for comparing stuff that
		// doesn't really have a representation.
		inputJSON   string
		errContains string
	}{
		{
			desc:        "empty",
			errContains: "empty input",
		},
		{
			desc:        "invalid source",
			line:        "111|ifmt",
			errContains: "invalid source",
		},
		{
			desc:        "invalid format",
			line:        "src|111",
			errContains: "invalid input format",
		},
		{
			desc:      "no modifiers",
			line:      "src|ifmt",
			inputJSON: `{"Source":{"Data":null},"Modifiers":null,"Format":{}}`,
		},
		{
			desc:      "one modifier",
			line:      "src|dbl|ifmt",
			inputJSON: `{"Source":{"Data":null},"Modifiers":[{"OptCloseErr":null}],"Format":{}}`,
		},
		{
			desc:      "many modifier",
			line:      "src|dbl|dbl|dbl|dbl|ifmt",
			inputJSON: `{"Source":{"Data":null},"Modifiers":[{"OptCloseErr":null},{"OptCloseErr":null},{"OptCloseErr":null},{"OptCloseErr":null}],"Format":{}}`,
		},
	}
	for _, c := range cases {
		parser := newTestParser()
		t.Run(c.desc, func(t *testing.T) {
			i, err := parser.parseInput(core.Context{}, c.line)
			if c.errContains == "" {
				assert.NilError(t, err)
				b, err := json.Marshal(i)
				assert.NilError(t, err)
				assert.Equal(t, c.inputJSON, string(b))
			} else {
				assert.ErrorContains(t, err, c.errContains)
			}
		})
	}
}

func TestParseOutput(t *testing.T) {
	cases := []struct {
		desc string
		line string
		// A hack, but easier for comparing stuff that
		// doesn't really have a representation.
		outputJSON  string
		errContains string
	}{
		{
			desc:        "empty",
			errContains: "empty output",
		},
		{
			desc:        "invalid format",
			line:        "111|snk",
			errContains: "invalid output format",
		},
		{
			desc:        "invalid sink",
			line:        "ofmt|111",
			errContains: "invalid sink",
		},
		{
			desc:       "no modifiers",
			line:       "ofmt|snk",
			outputJSON: `{"Format":{},"Modifiers":null,"Sink":{"OptCloseErr":null}}`,
		},
		{
			desc:       "one modifier",
			line:       "ofmt|dbl|snk",
			outputJSON: `{"Format":{},"Modifiers":[{"OptCloseErr":null}],"Sink":{"OptCloseErr":null}}`,
		},
		{
			desc:       "many modifier",
			line:       "ofmt|dbl|dbl|dbl|dbl|snk",
			outputJSON: `{"Format":{},"Modifiers":[{"OptCloseErr":null},{"OptCloseErr":null},{"OptCloseErr":null},{"OptCloseErr":null}],"Sink":{"OptCloseErr":null}}`,
		},
	}
	for _, c := range cases {
		parser := newTestParser()
		ctx := core.NewContext(&core.Config{}, &core.FlowConfig{}, nil)
		t.Run(c.desc, func(t *testing.T) {
			i, err := parser.parseOutput(ctx, c.line)
			if c.errContains == "" {
				assert.NilError(t, err)
				b, err := json.Marshal(i)
				assert.NilError(t, err)
				assert.Equal(t, c.outputJSON, string(b))
			} else {
				assert.ErrorContains(t, err, c.errContains)
			}
		})
	}
}
