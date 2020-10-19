package flow

// MetadataKey is a key for getting/settingx
// values in an instance of Metadata.
type MetadataKey int

const (
	// SourceID is the ID of the source.
	SourceID MetadataKey = iota
	// SourceName is the name of the source.
	SourceName
)

// Metadata holds additional data that
// may be useful at different parts of a flow.
type Metadata map[MetadataKey]string
