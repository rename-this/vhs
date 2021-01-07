package flow

import (
	"encoding/json"
	"testing"

	"github.com/rename-this/vhs/session"
	"gotest.tools/v3/assert"
)

func newTestParser() *Parser {
	return &Parser{
		Sources: map[string]SourceCtor{
			"src": newTestSource,
		},

		InputFormats: map[string]InputFormatCtor{
			"ifmt": newTestInputFormat,
		},

		OutputFormats: map[string]OutputFormatCtor{
			"ofmt": newTestOutputFormat,
		},

		Sinks: map[string]SinkCtor{
			"snk": func(_ session.Context) (Sink, error) {
				return &testSink{}, nil
			},
		},

		InputModifiers: map[string]InputModifierCtor{
			"dbl": func(_ session.Context) (InputModifier, error) {
				return &TestDoubleInputModifier{}, nil
			},
		},

		OutputModifiers: map[string]OutputModifierCtor{
			"dbl": func(_ session.Context) (OutputModifier, error) {
				return &TestDoubleOutputModifier{}, nil
			},
		},
	}
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
			flowJSON: `{"Input":{"Source":{},"Modifiers":[{},{},{},{}],"Format":{}},"Outputs":[{"Format":{},"Modifiers":[{},{},{},{}],"Sink":{}},{"Format":{},"Modifiers":[{},{},{},{}],"Sink":{}}]}`,
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
		ctx := session.NewContexts(&session.Config{}, &session.FlowConfig{}, nil)
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
			inputJSON: `{"Source":{},"Modifiers":null,"Format":{}}`,
		},
		{
			desc:      "one modifier",
			line:      "src|dbl|ifmt",
			inputJSON: `{"Source":{},"Modifiers":[{}],"Format":{}}`,
		},
		{
			desc:      "many modifier",
			line:      "src|dbl|dbl|dbl|dbl|ifmt",
			inputJSON: `{"Source":{},"Modifiers":[{},{},{},{}],"Format":{}}`,
		},
	}
	for _, c := range cases {
		parser := newTestParser()
		t.Run(c.desc, func(t *testing.T) {
			i, err := parser.parseInput(session.Context{}, c.line)
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
			outputJSON: `{"Format":{},"Modifiers":null,"Sink":{}}`,
		},
		{
			desc:       "one modifier",
			line:       "ofmt|dbl|snk",
			outputJSON: `{"Format":{},"Modifiers":[{}],"Sink":{}}`,
		},
		{
			desc:       "many modifier",
			line:       "ofmt|dbl|dbl|dbl|dbl|snk",
			outputJSON: `{"Format":{},"Modifiers":[{},{},{},{}],"Sink":{}}`,
		},
	}
	for _, c := range cases {
		parser := newTestParser()
		ctx := session.NewContexts(&session.Config{}, &session.FlowConfig{}, nil)
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
