---
title: "Format"
weight: 3
---
Format components apply structural formats to the data stream as it passes through them. For example, format components
may parse a datastream into HTTP requests and responses or format the data stream as JSON or another structured
encoding. Formats may exist in either the input or output chains of the `vhs` data flow. Input formats and output
formats must be implemented separately. Information on the input and output formats available in the current release of
`vhs` is available here: [input formats](/vhs/reference/#input-formats) and [output
formats](/vhs/reference/#output-formats)

## Developing a Format

Internally, the implementation of a format bears some resemblance to that of a [source](/vhs/architecture/source/),
with the added wrinkle of reading from an incoming channel.

Input formats must conform to the `InputFormat` interface as implemented in
[`core/input_formats.go`](https://github.com/rename-this/vhs/blob/main/core/input_format.go) and shown below:

```go
// InputFormat is an interface for formatting input
type InputFormat interface {
     Init(session.Context, middleware.Middleware, <-chan InputReader)
     Out() <-chan interface{}
}
```

An input format must read from a channel of
[`InputReader`](https://github.com/rename-this/vhs/blob/main/core/input_reader.go) passed to the `Init` function and
write formatted output to the `Out()` channel.

Output formats must conform to the `OutputFormat` interface as implemented in
[`core/output_format.go`](https://github.com/rename-this/vhs/blob/main/core/output_format.go) and shown below:

```go
// OutputFormat is an interface for formatting output
type OutputFormat interface {
     Init(session.Context, io.Writer)
     In() chan<- interface{}
}
```

The output format interface is essentially the reverse of the input format interface. An output format must read
`interface{}` from the `In()` channel and write bytes to the `io.Writer` passed to the `Init` function.

In both the input format and the output format, the `Init` function is responsible for the initialization and internal
processing of the format. It will be run in its own goroutine by the `vhs` flow infrastructure. It may use internal
goroutines, but all internal goroutines should clean up and exit upon receipt of a signal on `core.Context.Done()`.
Since formats read and write concurrently, avoidance of deadlocks and bottlenecks is critical. It may be useful to
employ two goroutines internally, one to read from the incoming channel and one to write to the outgoing location to
avoid contention and blocking.

Other considerations when developing formats are similar to other components:

* A format must provide a constructor. By convention this constructor should take an argument of type
`session.Context`. This constructor may be trivial.
* Formats may utilize command line arguments if necessary. These flags must be declared in `newRootCmd` in
[`cmd/vhs/main.go`](https://github.com/rename-this/vhs/blob/main/cmd/vhs/main.go). The values of these flags should be
stored in `FlowConfig` in the `session` package at
[`core/config.go`](https://github.com/rename-this/vhs/blob/main/core/config.go). These values will then be
available within your package as part of the `core.Context`.
* The connection between the input format and the output format is the only location in the `vhs` data flow that is not
strongly typed. Developers should consider this when designing formats and provide documentation indicating the
compatibility of their formats with other formats.

### Example

The `json` [input](/vhs/reference/#json) and [output formats](/vhs/reference/#json-1) are useful examples of input and
output format implementations. They are found in the `jsonx` package at
[`jsonx/jsonx.go`](https://github.com/rename-this/vhs/blob/main/jsonx/jsonx.go).

The `http` [input](/vhs/reference/#http) and output formats present more complex examples of formats, including
handling of middleware on the input format. They are implemented in the `httpx` package at
[`httpx/input_format.go`](https://github.com/rename-this/vhs/blob/main/httpx/input_format.go) and
[`httpx/output_format.go`](https://github.com/rename-this/vhs/blob/main/httpx/output_format.go).
