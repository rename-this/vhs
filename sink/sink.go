package sink

// Sink is a writable location for output.
type Sink interface {
	Init()
	Write(interface{}) error
	Flush() error
}
