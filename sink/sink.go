package sink

// Sink is a writable location for output.
type Sink interface {
	Write(interface{}) error
	Flush() error
}
