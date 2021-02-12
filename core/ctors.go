package core

type (
	// SourceCtor is a map of string to source constructors.
	SourceCtor func(Context) (Source, error)
	// InputModifierCtor is a map of string to input modifier constructors.
	InputModifierCtor func(Context) (InputModifier, error)
	// InputFormatCtor is a map of string to input format constructors.
	InputFormatCtor func(Context) (InputFormat, error)

	// OutputFormatCtor is a map of string to output format constructors.
	OutputFormatCtor func(Context) (OutputFormat, error)
	// OutputModifierCtor is a map of string to output modifier constructors.
	OutputModifierCtor func(Context) (OutputModifier, error)
	// SinkCtor is a map of string to sink constructors.
	SinkCtor func(Context) (Sink, error)
)
