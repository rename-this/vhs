package core

// Source is a data source that can be consumed
// by an input pipe.
type Source interface {
	Init(Context)
	Streams() <-chan InputReader
}
