package plugin

import (
	"fmt"
	stdplugin "plugin"

	"github.com/rename-this/vhs/flow"
	"github.com/rename-this/vhs/session"
)

const (
	sourcesFuncName        = "Sources"
	inputModifiersFuncName = "InputModifiers"
	inputFormatsFuncName   = "InputFormats"

	outputFormatsFuncName   = "OutputFormats"
	outputModifiersFuncName = "OutputModifiers"
	sinksFuncName           = "Sinks"
)

type (
	sourcesFuncType        = func() map[string]flow.SourceCtor
	inputModifiersFuncType = func() map[string]flow.InputModifierCtor
	inputFormatsFuncType   = func() map[string]flow.InputFormatCtor

	outputFormatsFuncType   = func() map[string]flow.OutputFormatCtor
	outputModifiersFuncType = func() map[string]flow.OutputModifierCtor
	sinksFuncType           = func() map[string]flow.SinkCtor
)

// Plugin loads flow components.
type Plugin struct {
	sp *stdplugin.Plugin
}

// Load loads a plugin from disk.
func Load(path string) (*Plugin, error) {
	p, err := stdplugin.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed open plugin: %w", err)
	}

	return &Plugin{
		sp: p,
	}, nil
}

// Summary summarizes the result of a plugn application.
type Summary struct {
	Replaced []string
}

// Apply adds and flow components from the plugin
// to the parser.
func (p *Plugin) Apply(ctx session.Context, parser *flow.Parser) (Summary, error) {
	var s Summary

	sourcesSymbol, err := p.sp.Lookup(sourcesFuncName)
	if err != nil {
		ctx.Logger.Debug().Msgf("failed to lookup %s: %v", sourcesFuncName, err)
	} else {
		sources, ok := sourcesSymbol.(sourcesFuncType)
		if !ok {
			return s, fmt.Errorf("failed to assert type of %s: %w", sourcesFuncName, err)
		}
		for name, ctor := range sources() {
			if parser.LoadSource(name, ctor) {
				s.Replaced = append(s.Replaced, name)
			}
		}
	}

	inputModifiersSymbol, err := p.sp.Lookup(inputModifiersFuncName)
	if err != nil {
		ctx.Logger.Debug().Msgf("failed to lookup %s: %v", inputModifiersFuncName, err)
	} else {
		inputModifiers, ok := inputModifiersSymbol.(inputModifiersFuncType)
		if !ok {
			return s, fmt.Errorf("failed to assert type of %s: %w", inputModifiersFuncName, err)
		}
		for name, ctor := range inputModifiers() {
			if parser.LoadInputModifier(name, ctor) {
				s.Replaced = append(s.Replaced, name)
			}
		}
	}

	inputFormatsSymbol, err := p.sp.Lookup(inputFormatsFuncName)
	if err != nil {
		ctx.Logger.Debug().Msgf("failed to lookup %s: %v", inputFormatsFuncName, err)
	} else {
		inputFormats, ok := inputFormatsSymbol.(inputFormatsFuncType)
		if !ok {
			return s, fmt.Errorf("failed to assert type of %s: %w", inputFormatsFuncName, err)
		}
		for name, ctor := range inputFormats() {
			if parser.LoadInputFormat(name, ctor) {
				s.Replaced = append(s.Replaced, name)
			}
		}
	}

	outputFormatsSymbol, err := p.sp.Lookup(outputFormatsFuncName)
	if err != nil {
		ctx.Logger.Debug().Msgf("failed to lookup %s: %v", outputFormatsFuncName, err)
	} else {
		outputFormats, ok := outputFormatsSymbol.(outputFormatsFuncType)
		if !ok {
			return s, fmt.Errorf("failed to assert type of %s: %w", outputFormatsFuncName, err)
		}
		for name, ctor := range outputFormats() {
			if parser.LoadOutputFormat(name, ctor) {
				s.Replaced = append(s.Replaced, name)
			}
		}
	}

	outputModifiersSymbol, err := p.sp.Lookup(outputModifiersFuncName)
	if err != nil {
		ctx.Logger.Debug().Msgf("failed to lookup %s: %v", outputModifiersFuncName, err)
	} else {
		outputModifers, ok := outputModifiersSymbol.(outputModifiersFuncType)
		if !ok {
			return s, fmt.Errorf("failed to assert type of %s: %w", outputModifiersFuncName, err)
		}
		for name, ctor := range outputModifers() {
			if parser.LoadOutputModifier(name, ctor) {
				s.Replaced = append(s.Replaced, name)
			}
		}
	}

	sinksSymbol, err := p.sp.Lookup(sinksFuncName)
	if err != nil {
		ctx.Logger.Debug().Msgf("failed to lookup %s: %v", sinksFuncName, err)
	} else {
		sinks, ok := sinksSymbol.(sinksFuncType)
		if !ok {
			return s, fmt.Errorf("failed to assert type of %s: %w", sinksFuncName, err)
		}
		for name, ctor := range sinks() {
			if parser.LoadSink(name, ctor) {
				s.Replaced = append(s.Replaced, name)
			}
		}
	}

	return s, nil
}
