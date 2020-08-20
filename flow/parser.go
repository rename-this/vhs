package flow

import (
	"os"
	"strings"

	"github.com/gramLabs/vhs/format"
	"github.com/gramLabs/vhs/gcs"
	"github.com/gramLabs/vhs/httpx"
	"github.com/gramLabs/vhs/modifier"
	"github.com/gramLabs/vhs/pipe"
	"github.com/gramLabs/vhs/session"
	"github.com/gramLabs/vhs/sink"
	"github.com/gramLabs/vhs/source"
	"github.com/gramLabs/vhs/tcp"

	"github.com/go-errors/errors"
)

const (
	// Separator is the character used separate flow parts.
	Separator = "|"
)

type (
	sourceCtor       func(*session.Context) (source.Source, error)
	inputFormatCtor  func(*session.Context) (format.Input, error)
	outputFormatCtor func(*session.Context) (format.Output, error)
	sinkCtor         func(*session.Context) (sink.Sink, error)
	readCloserCtor   func(*session.Context) (modifier.ReadCloser, error)
	writeCloserCtor  func(*session.Context) (modifier.WriteCloser, error)
)

// Parser parses text into a *flow.Flow
type Parser struct {
	sources       map[string]sourceCtor
	inputFormats  map[string]inputFormatCtor
	outputFormats map[string]outputFormatCtor
	sinks         map[string]sinkCtor
	readClosers   map[string]readCloserCtor
	writeClosers  map[string]writeCloserCtor
}

// DefaultParser is the default flow parser.
var DefaultParser = &Parser{
	sources: map[string]sourceCtor{
		"tcp": tcp.NewSource,
		"gcs": gcs.NewSource,
	},

	inputFormats: map[string]inputFormatCtor{
		"http": httpx.NewInputFormat,
	},

	outputFormats: map[string]outputFormatCtor{
		"har":     httpx.NewHAR,
		"json":    format.NewJSON,
		"jsonbuf": format.NewJSONBuffered,
	},

	sinks: map[string]sinkCtor{
		"gcs": gcs.NewSink,
		"stdout": func(_ *session.Context) (sink.Sink, error) {
			return os.Stdout, nil
		},
	},

	readClosers: map[string]readCloserCtor{
		"gzip": modifier.NewGzipReadCloser,
	},

	writeClosers: map[string]writeCloserCtor{
		"gzip": modifier.NewGzipWriteCloser,
	},
}

// Parse parses text into a flow.
func (p *Parser) Parse(ctx *session.Context, inputLine string, outputLines []string) (*Flow, error) {
	input, err := p.parseInput(ctx, inputLine)
	if err != nil {
		return nil, errors.Errorf("falied to parse input: %v", err)
	}

	var outputs pipe.Outputs
	for _, outputLine := range outputLines {
		o, err := p.parseOutput(ctx, outputLine)
		if err != nil {
			return nil, errors.Errorf("falied to parse outputs: %v", err)
		}
		outputs = append(outputs, o)
	}

	return &Flow{
		Input:   input,
		Outputs: outputs,
	}, nil
}

// parseInput parses an input line.
// Examples;
// 		tcp|http
// 		gcs|gzip|json
// The first part is expected to be a valid source, the last is expected
// to be a valid input format. Any parts in the middle are modifiers.
func (p *Parser) parseInput(ctx *session.Context, line string) (*pipe.Input, error) {
	if line == "" {
		return nil, errors.New("empty input")
	}

	var (
		s   source.Source
		f   format.Input
		rcs modifier.ReadClosers
		err error

		parts = strings.Split(line, Separator)
	)

	sPart := parts[0]
	sCtor, ok := p.sources[sPart]
	if !ok {
		return nil, errors.Errorf("invalid source: %s", sPart)
	}
	s, err = sCtor(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to create source: %v", err)
	}

	fPart := parts[len(parts)-1]
	fCtor, ok := p.inputFormats[fPart]
	if !ok {
		return nil, errors.Errorf("invalid input format: %s", fPart)
	}
	f, err = fCtor(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to create input format: %v", err)
	}

	for _, rcPart := range parts[1 : len(parts)-1] {
		rcCtor, ok := p.readClosers[rcPart]
		if !ok {
			return nil, errors.Errorf("invalid modifier: %s", fPart)
		}
		rc, err := rcCtor(ctx)
		if err != nil {
			return nil, errors.Errorf("failed to create modifier: %v", err)
		}
		rcs = append(rcs, rc)
	}

	return pipe.NewInput(f, s, rcs), nil
}

// parseOutput parses an output line.
// Examples;
// 		json|gzip|gcs
// 		http|har
// The first part is expected to be a valid output format, the last is expected
// to be a valid sink. Any parts in the middle are modifiers.
func (p *Parser) parseOutput(ctx *session.Context, line string) (*pipe.Output, error) {
	if line == "" {
		return nil, errors.New("empty output")
	}

	var (
		f   format.Output
		s   sink.Sink
		wcs modifier.WriteClosers
		err error

		parts = strings.Split(line, Separator)
	)

	fPart := parts[0]
	fCtor, ok := p.outputFormats[fPart]
	if !ok {
		return nil, errors.Errorf("invalid output format: %s", fPart)
	}
	f, err = fCtor(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to create output format: %v", err)
	}

	sPart := parts[len(parts)-1]
	sCtor, ok := p.sinks[sPart]
	if !ok {
		return nil, errors.Errorf("invalid sink: %s", sPart)
	}
	s, err = sCtor(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to create sink: %v", err)
	}

	for _, wcPart := range parts[1 : len(parts)-1] {
		wcCtor, ok := p.writeClosers[wcPart]
		if !ok {
			return nil, errors.Errorf("invalid modifier: %s", fPart)
		}
		wc, err := wcCtor(ctx)
		if err != nil {
			return nil, errors.Errorf("failed to create modifier: %v", err)
		}
		wcs = append(wcs, wc)
	}

	return pipe.NewOutput(f, s, wcs), nil
}
