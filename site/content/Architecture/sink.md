---
title: "Sink"
weight: 4

---

A sink component is the termination of an output chain in the `vhs` data flow. Each output chain must end with a sink.
A sink provides a way for data to leave `vhs`, such as by writing to the standard output, a local file, or a network
location, among others. Information on the sinks available in the current release of `vhs` can be found here:
[sinks](/vhs/reference/#sinks).

## Developing a Sink

Sinks are the simplest components in `vhs`. Sinks must conform to the the interface defined in
[`flow/sink.go`](https://github.com/rename-this/vhs/blob/main/flow/sink.go) as shown below:

```go
// Sink is a writable location for output.
type Sink io.WriteCloser
```

### Example

Since a sink is simply an `io.WriteCloser`, any existing types that conform to this interface can be used a sink. This
means that a simple sink can be implemented in only a handful of lines of code, as with the [`stdout`
sink](/vhs/reference/#stdout), which is implemented in
[`cmd/vhs/main.go`](https://github.com/rename-this/vhs/blob/main/cmd/vhs/main.go#L262)
in the default parser definition. Since it is so short, it is recreated here:

```go
p.LoadSink("stdout", func(_ session.Context) (flow.Sink, error) {
    return os.Stdout, nil
})
```

Since `os.Stdout` conforms to `io.WriteCloser`, the `stdout` sink can be implemented as a one-line anonymous function.

The `gcs` sink presents a more complex, real-world example of writing to cloud storage. It is implemented in
[`gcs/gcs.go`](https://github.com/rename-this/vhs/blob/main/gcs/gcs.go).
