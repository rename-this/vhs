---
title: "Source"
weight: 1
---

A source component is the origin of a data flow in `vhs`. Source components create the initial data stream that is
processed and modified by the other components in the data flow. A source might read from a local file, a cloud storage
location, or capture network data. Information on the sources available in the current release of `vhs` is available on
the reference page here: [sources](/vhs/reference/#sources).

## Developing a Source

There are several architectural requirements that must be considered when developing a source. Most importantly,
sources must conform to the the interface defined in
[`flow/source.go`](https://github.com/rename-this/vhs/blob/main/flow/source.go) and shown below:

```go

// Source is a data source that can be consumed
// by an input pipe.
type Source interface {
    Init(session.Context)
    Streams() <-chan InputReader
}
```

The `Init` function should handle the initialization and concurrent operation of the source. It will be run in its own
goroutine by the `vhs` flow infrastructure. `vhs` does not use buffered channels, and all components are expected to
run concurrently, so care should be taken to avoid internal deadlocks or contention that could cause blocks or
bottlenecks. The Init function should clean up and exit upon receiving a signal on `session.Context.Done()`.

Other important considerations include:

* A source must provide a constructor. By convention this constructor should accept an argument of type
`session.Context`. This constructor may be trivial.
* Sources may declare command line flags in `newRootCmd()` in
[`cmd/vhs/main.go`](https://github.com/rename-this/vhs/blob/main/cmd/vhs/main.go). The values of these flags should be
stored in `FlowConfig` in the `session` package at
[`session/config.go`](https://github.com/rename-this/vhs/blob/main/session/config.go). These values will then be
available within your package as part of the `session.Context`.

### Example

The [`file` source](/vhs/reference/#file) is a good example implementation of a source. This source reads data from a
file and emits it as a stream of bytes. It is implemented in the `file` package at
[`file/file.go`](https://github.com/rename-this/vhs/blob/main/file/file.go)
