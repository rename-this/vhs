package flow

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/rename-this/vhs/core"
)

const (
	// Separator is the character used separate flow parts.
	Separator = "|"
)

// NewParser creates a new parser.
func NewParser() *Parser {
	return &Parser{
		sources:         make(map[string]core.SourceCtor),
		inputModifiers:  make(map[string]core.InputModifierCtor),
		inputFormats:    make(map[string]core.InputFormatCtor),
		outputFormats:   make(map[string]core.OutputFormatCtor),
		outputModifiers: make(map[string]core.OutputModifierCtor),
		sinks:           make(map[string]core.SinkCtor),
	}
}

// Parser parses text into a *flow.Flow
type Parser struct {
	mu sync.RWMutex

	sources        map[string]core.SourceCtor
	inputModifiers map[string]core.InputModifierCtor
	inputFormats   map[string]core.InputFormatCtor

	outputFormats   map[string]core.OutputFormatCtor
	outputModifiers map[string]core.OutputModifierCtor
	sinks           map[string]core.SinkCtor
}

// LoadSource loads a new source and returns a value indicating
// whether the value replaced a previous entry.
func (p *Parser) LoadSource(name string, ctor core.SourceCtor) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, replaced := p.sources[name]
	p.sources[name] = ctor
	return replaced
}

// LoadInputModifier loads a new input modifier and returns a value
// inidicating whether the value replaced a previous entry.
func (p *Parser) LoadInputModifier(name string, ctor core.InputModifierCtor) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, replaced := p.inputModifiers[name]
	p.inputModifiers[name] = ctor
	return replaced
}

// LoadInputFormat loads a new input format and returns a value
// indicating whether the value replaced a previous entry.
func (p *Parser) LoadInputFormat(name string, ctor core.InputFormatCtor) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, replaced := p.inputFormats[name]
	p.inputFormats[name] = ctor
	return replaced
}

// LoadOutputFormat loads a new output format and returns a value
// indicating whether the value replaced a previous entry.
func (p *Parser) LoadOutputFormat(name string, ctor core.OutputFormatCtor) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, replaced := p.outputFormats[name]
	p.outputFormats[name] = ctor
	return replaced
}

// LoadOutputModifier loads a new output modifier and returns a value
// indicating whether the value replaced a previous entry.
func (p *Parser) LoadOutputModifier(name string, ctor core.OutputModifierCtor) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, replaced := p.outputModifiers[name]
	p.outputModifiers[name] = ctor
	return replaced
}

// LoadSink loads a new sink and returns a value indicatin whether
// the value replaced a previous entry.
func (p *Parser) LoadSink(name string, ctor core.SinkCtor) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, replaced := p.sinks[name]
	p.sinks[name] = ctor
	return replaced
}

// Parse parses text into a flow.
func (p *Parser) Parse(ctx core.Context, inputLine string, outputLines []string) (*Flow, error) {
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
func (p *Parser) parseInput(ctx core.Context, line string) (*Input, error) {
	if line == "" {
		return nil, errors.New("empty input")
	}

	var (
		s    core.Source
		f    core.InputFormat
		mods core.InputModifiers
		err  error

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
		mods = append(mods, rc)
	}

	return NewInput(s, mods, f), nil
}

// parseOutput parses an output line.
// Examples;
// 		json|gzip|gcs
// 		http|har
// The first part is expected to be a valid output format, the last is expected
// to be a valid sink. Any parts in the middle are modifiers.
func (p *Parser) parseOutput(ctx core.Context, line string) (*Output, error) {
	if line == "" {
		return nil, errors.New("empty output")
	}

	var (
		f    core.OutputFormat
		s    core.Sink
		mods core.OutputModifiers
		err  error

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
		mods = append(mods, wc)
	}

	return NewOutput(f, mods, s), nil
}
