package core

// InputFormat is an interface for formatting input.
type InputFormat interface {
	Init(Context, Middleware, <-chan InputReader)
	Out() <-chan interface{}
}
