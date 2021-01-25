package flow

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/rename-this/vhs/session"
)

const (
	// Separator is the character used separate flow parts.
	Separator = "|"
)

type (
	// SourceCtor is a map of string to source constructors.
	SourceCtor func(session.Context) (Source, error)
	// InputModifierCtor is a map of string to input modifier constructors.
	InputModifierCtor func(session.Context) (InputModifier, error)
	// InputFormatCtor is a map of string to input format constructors.
	InputFormatCtor func(session.Context) (InputFormat, error)

	// OutputFormatCtor is a map of string to output format constructors.
	OutputFormatCtor func(session.Context) (OutputFormat, error)
	// OutputModifierCtor is a map of string to output modifier constructors.
	OutputModifierCtor func(session.Context) (OutputModifier, error)
	// SinkCtor is a map of string to sink constructors.
	SinkCtor func(session.Context) (Sink, error)
)

// NewParser creates a new parser.
func NewParser() *Parser {
	return &Parser{
		sources:         make(map[string]SourceCtor),
		inputModifiers:  make(map[string]InputModifierCtor),
		inputFormats:    make(map[string]InputFormatCtor),
		outputFormats:   make(map[string]OutputFormatCtor),
		outputModifiers: make(map[string]OutputModifierCtor),
		sinks:           make(map[string]SinkCtor),
	}
}

// Parser parses text into a *flow.Flow
type Parser struct {
	mu sync.RWMutex

	sources        map[string]SourceCtor
	inputModifiers map[string]InputModifierCtor
	inputFormats   map[string]InputFormatCtor

	outputFormats   map[string]OutputFormatCtor
	outputModifiers map[string]OutputModifierCtor
	sinks           map[string]SinkCtor
}

// LoadSource loads a new source and returns a value indicating
// whether the value replaced a previous entry.
func (p *Parser) LoadSource(name string, ctor SourceCtor) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, replaced := p.sources[name]
	p.sources[name] = ctor
	return replaced
}

// LoadInputModifier loads a new input modifier and returns a value
// inidicating whether the value replaced a previous entry.
func (p *Parser) LoadInputModifier(name string, ctor InputModifierCtor) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, replaced := p.inputModifiers[name]
	p.inputModifiers[name] = ctor
	return replaced
}

// LoadInputFormat loads a new input format and returns a value
// indicating whether the value replaced a previous entry.
func (p *Parser) LoadInputFormat(name string, ctor InputFormatCtor) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, replaced := p.inputFormats[name]
	p.inputFormats[name] = ctor
	return replaced
}

// LoadOutputFormat loads a new output format and returns a value
// indicating whether the value replaced a previous entry.
func (p *Parser) LoadOutputFormat(name string, ctor OutputFormatCtor) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, replaced := p.outputFormats[name]
	p.outputFormats[name] = ctor
	return replaced
}

// LoadOutputModifier loads a new output modifier and returns a value
// indicating whether the value replaced a previous entry.
func (p *Parser) LoadOutputModifier(name string, ctor OutputModifierCtor) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, replaced := p.outputModifiers[name]
	p.outputModifiers[name] = ctor
	return replaced
}

// LoadSink loads a new sink and returns a value indicatin whether
// the value replaced a previous entry.
func (p *Parser) LoadSink(name string, ctor SinkCtor) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, replaced := p.sinks[name]
	p.sinks[name] = ctor
	return replaced
}

// Parse parses text into a flow.
func (p *Parser) Parse(ctx session.Context, inputLine string, outputLines []string) (*Flow, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	input, err := p.parseInput(ctx, inputLine)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input: %v", err)
	}

	var outputs []*Output
	for _, outputLine := range outputLines {
		o, err := p.parseOutput(ctx, outputLine)
		if err != nil {
			return nil, fmt.Errorf("failed to parse outputs: %v", err)
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
func (p *Parser) parseInput(ctx session.Context, line string) (*Input, error) {
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
	sCtor, ok := p.sources[sPart]
	if !ok {
		return nil, fmt.Errorf("invalid source: %s", sPart)
	}
	s, err = sCtor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create source: %v", err)
	}

	fPart := parts[len(parts)-1]
	fCtor, ok := p.inputFormats[fPart]
	if !ok {
		return nil, fmt.Errorf("invalid input format: %s", fPart)
	}
	f, err = fCtor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create input format: %v", err)
	}

	for _, rcPart := range parts[1 : len(parts)-1] {
		rcCtor, ok := p.inputModifiers[rcPart]
		if !ok {
			return nil, fmt.Errorf("invalid modifier: %s", fPart)
		}
		rc, err := rcCtor(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create modifier: %v", err)
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
func (p *Parser) parseOutput(ctx session.Context, line string) (*Output, error) {
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
	fCtor, ok := p.outputFormats[fPart]
	if !ok {
		return nil, fmt.Errorf("invalid output format: %s", fPart)
	}
	f, err = fCtor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create output format: %v", err)
	}

	sPart := parts[len(parts)-1]
	sCtor, ok := p.sinks[sPart]
	if !ok {
		return nil, fmt.Errorf("invalid sink: %s", sPart)
	}
	s, err = sCtor(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create sink: %v", err)
	}

	for _, wcPart := range parts[1 : len(parts)-1] {
		wcCtor, ok := p.outputModifiers[wcPart]
		if !ok {
			return nil, fmt.Errorf("invalid modifier: %s", fPart)
		}
		wc, err := wcCtor(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to create modifier: %v", err)
		}
		mos = append(mos, wc)
	}

	return NewOutput(f, mos, s), nil
}
