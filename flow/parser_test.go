package flow

import (
	"encoding/json"
	"testing"

	"github.com/gramLabs/vhs/config"
	"github.com/gramLabs/vhs/modifier"
	"github.com/gramLabs/vhs/session"
	"github.com/gramLabs/vhs/sink"
	"github.com/gramLabs/vhs/testhelper"
	"gotest.tools/assert"
)

var (
	testParser = &Parser{
		Sources: map[string]SourceCtor{
			"src": testhelper.NewSource,
		},

		InputFormats: map[string]InputFormatCtor{
			"ifmt": testhelper.NewInputFormat,
		},

		OutputFormats: map[string]OutputFormatCtor{
			"ofmt": testhelper.NewOutputFormat,
		},

		Sinks: map[string]SinkCtor{
			"snk": func(_ *session.Context) (sink.Sink, error) {
				return &testhelper.Sink{}, nil
			},
		},

		ReadClosers: map[string]ReadCloserCtor{
			"dbl": func(_ *session.Context) (modifier.ReadCloser, error) {
				return &testhelper.DoubleInput{}, nil
			},
		},

		WriteClosers: map[string]WriteCloserCtor{
			"dbl": func(_ *session.Context) (modifier.WriteCloser, error) {
				return &testhelper.DoubleOutput{}, nil
			},
		},
	}
	ifmt, _ = testhelper.NewInputFormat(nil)
	src, _  = testhelper.NewSource(nil)
	dblIn   = &testhelper.DoubleInput{}
	dblOut  = &testhelper.DoubleOutput{}
)

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
			inputJSON: `{"Source":{},"Format":{},"Modifiers":null}`,
		},
		{
			desc:      "one modifier",
			line:      "src|dbl|ifmt",
			inputJSON: `{"Source":{},"Format":{},"Modifiers":[{}]}`,
		},
		{
			desc:      "many modifier",
			line:      "src|dbl|dbl|dbl|dbl|ifmt",
			inputJSON: `{"Source":{},"Format":{},"Modifiers":[{},{},{},{}]}`,
		},
	}
	for _, c := range cases {
		t.Run(c.desc, func(t *testing.T) {
			i, err := testParser.parseInput(nil, c.line)
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
			outputJSON: `{"Format":{},"Sink":{},"Modifiers":null}`,
		},
		{
			desc:       "one modifier",
			line:       "ofmt|dbl|snk",
			outputJSON: `{"Format":{},"Sink":{},"Modifiers":[{}]}`,
		},
		{
			desc:       "many modifier",
			line:       "ofmt|dbl|dbl|dbl|dbl|snk",
			outputJSON: `{"Format":{},"Sink":{},"Modifiers":[{},{},{},{}]}`,
		},
	}
	for _, c := range cases {
		ctx, _, _ := session.NewContexts(&config.Config{}, nil)
		t.Run(c.desc, func(t *testing.T) {
			i, err := testParser.parseOutput(ctx, c.line)
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
