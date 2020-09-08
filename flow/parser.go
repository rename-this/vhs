package flow

import (
	"strings"

	"github.com/gramLabs/vhs/session"
	"github.com/pkg/errors"
)

const (
	// Separator is the character used separate flow parts.
	Separator = "|"
)

type (
	// SourceCtor is a map of string to source constructors.
	SourceCtor func(*session.Context) (Source, error)
	// InputModifierCtor is a map of string to input modifier constructors.
	InputModifierCtor func(*session.Context) (InputModifier, error)
	// InputFormatCtor is a map of string to input format constructors.
	InputFormatCtor func(*session.Context) (InputFormat, error)

	// OutputFormatCtor is a map of string to output format constructors.
	OutputFormatCtor func(*session.Context) (OutputFormat, error)
	// OutputModifierCtor is a map of string to output modifier constructors.
	OutputModifierCtor func(*session.Context) (OutputModifier, error)
	// SinkCtor is a map of string to sink constructors.
	SinkCtor func(*session.Context) (Sink, error)
)

// Parser parses text into a *flow.Flow
type Parser struct {
	Sources        map[string]SourceCtor
	InputModifiers map[string]InputModifierCtor
	InputFormats   map[string]InputFormatCtor

	OutputFormats   map[string]OutputFormatCtor
	OutputModifiers map[string]OutputModifierCtor
	Sinks           map[string]SinkCtor
}

// Parse parses text into a flow.
func (p *Parser) Parse(ctx *session.Context, inputLine string, outputLines []string) (*Flow, error) {
	input, err := p.parseInput(ctx, inputLine)
	if err != nil {
		return nil, errors.Errorf("failed to parse input: %v", err)
	}

	var outputs Outputs
	for _, outputLine := range outputLines {
		o, err := p.parseOutput(ctx, outputLine)
		if err != nil {
			return nil, errors.Errorf("failed to parse outputs: %v", err)
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
func (p *Parser) parseInput(ctx *session.Context, line string) (*Input, error) {
	if line == "" {
		return nil, errors.New("empty input")
	}

	var (
		s   Source
		f   InputFormat
		mis InputModifiers
		err error

		parts = strings.Split(line, Separator)
	)

	sPart := parts[0]
	sCtor, ok := p.Sources[sPart]
	if !ok {
		return nil, errors.Errorf("invalid source: %s", sPart)
	}
	s, err = sCtor(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to create source: %v", err)
	}

	fPart := parts[len(parts)-1]
	fCtor, ok := p.InputFormats[fPart]
	if !ok {
		return nil, errors.Errorf("invalid input format: %s", fPart)
	}
	f, err = fCtor(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to create input format: %v", err)
	}

	for _, rcPart := range parts[1 : len(parts)-1] {
		rcCtor, ok := p.InputModifiers[rcPart]
		if !ok {
			return nil, errors.Errorf("invalid modifier: %s", fPart)
		}
		rc, err := rcCtor(ctx)
		if err != nil {
			return nil, errors.Errorf("failed to create modifier: %v", err)
		}
		mis = append(mis, rc)
	}

	return NewInput(s, mis, f), nil
}

// parseOutput parses an output line.
// Examples;
// 		json|gzip|gcs
// 		http|har
// The first part is expected to be a valid output format, the last is expected
// to be a valid sink. Any parts in the middle are modifiers.
func (p *Parser) parseOutput(ctx *session.Context, line string) (*Output, error) {
	if line == "" {
		return nil, errors.New("empty output")
	}

	var (
		f   OutputFormat
		s   Sink
		mos OutputModifiers
		err error

		parts = strings.Split(line, Separator)
	)

	fPart := parts[0]
	fCtor, ok := p.OutputFormats[fPart]
	if !ok {
		return nil, errors.Errorf("invalid output format: %s", fPart)
	}
	f, err = fCtor(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to create output format: %v", err)
	}

	sPart := parts[len(parts)-1]
	sCtor, ok := p.Sinks[sPart]
	if !ok {
		return nil, errors.Errorf("invalid sink: %s", sPart)
	}
	s, err = sCtor(ctx)
	if err != nil {
		return nil, errors.Errorf("failed to create sink: %v", err)
	}

	for _, wcPart := range parts[1 : len(parts)-1] {
		wcCtor, ok := p.OutputModifiers[wcPart]
		if !ok {
			return nil, errors.Errorf("invalid modifier: %s", fPart)
		}
		wc, err := wcCtor(ctx)
		if err != nil {
			return nil, errors.Errorf("failed to create modifier: %v", err)
		}
		mos = append(mos, wc)
	}

	return NewOutput(f, mos, s), nil
}
