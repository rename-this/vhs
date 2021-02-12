---
title: "Modifier"
weight: 2
---

A modifier component operates on an unformatted, unstructured stream of bytes as it passes through `vhs`. Modifier
components are well suited to implementing compression/decompression algorithms, as an example. Modifiers may exist in
either the input or output chains of the `vhs` data flow. Input modifiers and output modifiers must be implemented
separately. Information on the input and output modifiers available in the current release of `vhs` is available here:
[input modifiers](/vhs/reference/#input-modifiers) and [output modifiers](/vhs/reference/#output-modifiers)

## Developing a Modifier

Modifiers are implemented in a slightly different fashion when compared to other `vhs` components. They do not have an
`Init` function, and are instead implemented as wrappers around an `InputReader` or `OutputWriter`.
Input modifiers must conform to the `InputModifier` interface defined in
[`core/input_modifier.go`](https://github.com/rename-this/vhs/blob/main/core/input_modifiers.go), which is shown below:

```go
// InputModifier wraps an InputReader.
type InputModifier interface {
     Wrap(InputReader) (InputReader, error)
}
```

Output modifiers must conform to the `OutputModifier` interface defined in
[`core/output_modifiers.go`](https://github.com/rename-this/vhs/blob/main/core/output_modifiers.go), which is shown below:

```go
// OutputModifier wraps an OtputWriter.
type OutputModifier interface {
     Wrap(OutputWriter) (OutputWriter, error)
}
```

Internally, the implementation of a modifier usually entails the creation of a new type that conforms to the requisite
interface (`InputReader` or `OutputWriter`). This the compression or encoding performed by the modifier will be
implemented in the methods of this new type, and the `Wrap` function required by the modifier interfaces seen above
will return this new type.

Other considerations when developing a modifier include:

* A modifier must provide a constructor. By convention this constructor should take an argument of type
`session.Context`. This constructor may be trivial.
* Modifiers may utilize command line arguments if necessary. These flags must be declared in `newRootCmd` in
[`cmd/vhs/main.go`](https://github.com/rename-this/vhs/blob/main/cmd/vhs/main.go). The values of these flags should be
stored in `FlowConfig` in the `session` package at
[`core/config.go`](https://github.com/rename-this/vhs/blob/main/core/config.go). These values will then be
available within your package as part of the `core.Context`.
* If a modifier uses any internal goroutines, these must clean up and exit upon receiving a signal on the
`core.Context.Done()` channel.

### Example

The `gzip` [input](/vhs/reference/#gzip) and [output modifiers](/vhs/reference/#gzip-1) provided by `vhs` are useful
examples of the implementation of modifiers. They are implemented as wrappers around the `gzip.Reader` and
`gzip.Writer` packages from the Go standard library package `compress/gzip`. Many useful modifiers can be implemented
in a similar fashion by wrapping other libraries. The implementation of the `gzip` modifiers is in
[`gzipx/gzipx.go`](https://github.com/rename-this/vhs/blob/main/gzipx/gzipx.go)
