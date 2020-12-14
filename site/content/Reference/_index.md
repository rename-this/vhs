---
title: "Reference"
linkTitle: "Reference"
weight: 3
description: >
    Usage and command line flags reference.
---
## Introduction
`vhs` is designed for flexibility and operates on the concept of a data flow that originates with a [source](#sources)
and terminates with one or more [sinks](#sinks). Sources may capture network data, read data from files, etc. Sinks may
write data to cloud or local storage, standard output, or send data to other destinations. Along the way, data may pass
through a series of input [modifiers](#input-modifiers) and [formats](#input-formats) and output 
[modifiers](#output-modifiers) and [formats](#output-formats) that transform the data. For more information on the 
technical implementation of `vhs`, see the [architecture overview](/vhs/architecture/).

## Example `vhs` Command
```./vhs --inputs "tcp|http" --outputs "json|stdout" --address 0.0.0.0:8080 --capture-response```

This command captures two-way TCP data from the network on address `0.0.0.0` and port `8080`, extracts HTTP requests
and responses from the TCP data, formats them as JSON, and prints them to the standard output.

## Specifying Inputs and Outputs
The core command line flags for `vhs` are focussed on defining the data flow that `vhs` will use for a recording/replay
session. Inputs and outputs are specified in terms of a simple domain specific language that will be detailed in
the following sections.

### Inputs
```--input "<source|modifier(s)|format>"```

Inputs are specified in a pipe-delimited (`|`), double-quoted string following the `--input` command line flag. 
The first element in the string must specify a [source](#sources), followed by a pipe character, followed by zero or
more [modifiers](#input-modifiers) separated by pipe characters, followed by another pipe character, and ending with an
input [format](#input-formats) specifier. The specified source originates a data stream that is modified by the 
specified modifiers and then formatted, or interpreted by the specified input format.

In the example command given above, the input specifier is `--inputs "tcp|http"` where `tcp` specifies the TCP source
and `http` specifies the HTTP input format. This example does not use any input modifiers. 

The next sections will detail the currently available [input sources](#input-sources), [modifiers](#input-modifiers),
and [formats](#input-formats) in `vhs`. Each source, modifier, and format may require additional configuration in the
form of additional command line flags. These will be described where applicable.

Only one input definition can be specified in a `vhs` session.

#### Sources
The following sources are currently available:
* `tcp`
* `file`
* `gcs` (Google cloud storage)
* `s3compat` (S3 compatible cloud storage)

##### `tcp`
The `tcp` source captures live TCP/IP network data. It uses the following additional command line flags for
configuration:
* `--address <ip address:port>` Required. Specifies the address and port on which `vhs` will listen.
* `--capture-response` Optional. If set, `vhs` captures requests and responses (2-way traffic).

##### `file`
The `file` source reads data from a file on the local filesystem. It requires the following command line flag
for configuration.
* `--input_file <path to input file>` Required. Specifies the path to the input file to be read.

##### `gcs`
The `gcs` source reads data from a Google Cloud Storage object. It requires the following command line flags for
configuration. Note that the GCS source also requires Google Cloud authentication credentials to be present
on the machine or in the container where `vhs` is run. For more information on GCS authentication, see Google's
documentation [here](https://cloud.google.com/docs/authentication/production).
* `--gcs-bucket-name <GCS bucket name>` Required. Name of bucket that contains the object to be read.
* `--gcs-object-name <object name>` Required. Name of object to be read.

Note that this source also requires a JSON key file containing Google Cloud authentication credentials.

##### `s3compat`
The `s3compat` source reads from an object in an S3-compatible cloud storage location. It requires the following command
line flags for configuration.
* `--s3-compat-access-key <access key>` Required. Access key for S3 compatible storage.
* `--s3-compat-secret-key <secret key>` Required. Secret key for S3 compatible storage.
* `--s3-compat-token <token>` Required. Session token for S3 compatible storage.
* `--s3-compat-secure` Optional. This flag specifies encrypted transport (HTTPS). Default is `true`.
* `--s3-compat-endpoint <S3 URL>` Required. URL for S3-compatible storage.
* `--s3-compat-bucket-name <bucket name>` Required. Name of bucket that contains the object to be read.
* `--s3-compat-object-name <object name>` Required. Name of object to be read.

#### Input Modifiers
The following input modifiers are currently available in `vhs`:
* `gzip`

##### `gzip`
The `gzip` input modifier uncompresses data that has been compressed in the gzip format. It is primarily for use with
the `file`, `gcs`, and `s3compat` sources, enabling the reading of compressed files.

#### Input Formats
The following input formats are currently available in `vhs`:
* `http`
* `json`

##### `http`
The `http` input format decodes the incoming data stream into HTTP requests and responses. This format is primarily
intended for use with the [`tcp` source](#tcp).

##### `json`
The `json` input format interprets the incoming data stream as JSON. It is primarily intended for use with the 
[`file`](#file)
and cloud storage sources ([`gcs`](#gcs) and [`s3compat`](#s3compat)) for processing data stored in a JSON file.

### Outputs
```--output "<format|modifier(s)|sink>"```

Outputs are specified in a pipe-delimited (`|`), double-quoted string following the `--output` command line flag.
The first element in the string must specify an [output format](#output-formats), followed by a pipe character, followed
by zero or more [modifiers](#output-modifiers) separated by pipe characters, followed by another pipe character, and 
ending with a [sink](#sinks) specifier. The output chain works similarly to the input chain. The input format receives
the data stream from the end of the input chain and formats or interprets the data. This data can then be modified by
an output modifier before it leaves `vhs` through a sink. 

In the example command given above, the output specifier is `--outputs "json|stdout"` where `json` specifies the JSON
output format and `stdout` specifies the standard output sink. This example does not use any output modifiers.

The next sections will detail the currently available [output format](#output-formats), [modifiers](#output-modifiers),
and [sink](#sinks) in `vhs`. Each format, modifier, and sink may require additional configuration in the form of 
additional command line flags. These will be described where applicable.

`vhs` supports an arbitrary number of outputs for any given session. Each output will receive the same data from
the input chain.

#### Output Formats
The following output formats are currently available in `vhs`:
* `har` (HTTP archive)
* `json`

##### `har`
The `har` output format receives incoming data in the form of a stream of HTTP requests and responses and encodes
it into the HTTP Archive (HAR) format. The output of this format can be saved to cloud storage or printed to standard
output depending on the [sink](#sinks) chosen by the user. For more information on the HTTP Archive format see
[the specification](http://www.softwareishard.com/blog/har-12-spec/).

##### `json`
The `json` output format receives incoming data in the form of a stream of HTTP requests and responses and serializes
those requests and responses to the JSON format. The output of this format can be saved to cloud storage or printed to 
standard output depending on the [sink](#sinks) chosen by the user.

#### Output Modifiers
The following output modifiers are currently available in `vhs`:
* `gzip`

##### `gzip`
The `gzip` output modifier compresses the data passing through it into the gzip format. This can be used in conjunction
with the [`stdout`](#stdout) or cloud storage ([`gcs`](#gcs-1) or [`s3compat`](#s3compat-1)) sinks to save compressed
output data from `vhs`.

#### Sinks
The following sinks are currently available in `vhs`:
* `gcs` (Google cloud storage)
* `s3compat` (S3-compatible cloud storage)
* `stdout`
* `discard`

##### `gcs`
The `gcs` sink writes data to a Google Cloud Storage object. It requires the following command line flags for
configuration. Note that the GCS sink also requires Google Cloud authentication credentials to be present
on the machine or in the container where `vhs` is run. For more information on GCS authentication, see Google's 
documentation [here](https://cloud.google.com/docs/authentication/production).
* `--gcs-bucket-name <GCS bucket name>` Required. Bucket name that contains the GCS object to be written to.
* `--gcs-object-name <object name>` Required. Name of object to be read.

##### `s3compat`
The `s3compat` sink writes to an object in an S3-compatible cloud storage location. It requires the following command
line flags for configuration.
* `--s3-compat-access-key <access key>` Required. Access key for S3 compatible storage.
* `--s3-compat-secret-key <secret key>` Required. Secret key for S3 compatible storage.
* `--s3-compat-token <token>` Required. Session token for S3 compatible storage.
* `--s3-compat-secure` Optional. This flag specifies encrypted transport (HTTPS). Default is `true`.
* `--s3-compat-endpoint <S3 URL>` Required. URL for S3-compatible storage.
* `--s3-compat-bucket-name <bucket name>` Required. Name of bucket that contains the object to be written.
* `--s3-compat-object-name <object name>` Required. Name of object to be written.

##### `stdout`
The `stdout` sink writes the data stream it receives to the standard output. This sink can be used in conjunction with 
shell redirection to save the output of `vhs` to a file on the local filesystem.

##### `discard`
The `discard` sink silently discards the data that is sent to it.

## Middleware
```--middleware <path to middleware executable>```

`vhs` optionally supports the use of user-supplied middleware to modify data as it passes through `vhs`. As an example,
user supplied middleware could be utilized to remove sensitive user credentials from recorded HTTP data before saving
it to cloud storage. Middleware, if used, is placed in the `vhs` data flow in the output chain between the output format
and the output modifiers. It is implemented as a separate binary that will receive formatted data on the standard input
and must write modified data on the standard output. A simple example middleware can be found 
[here](https://github.com/rename-this/vhs/blob/main/testdata/http_middleware.bash) in the `vhs` repository.

## Prometheus metrics 
```--prometheus-address <ip adddress:port>```

`vhs` supports calculating metrics on HTTP exchanges captured live from the network. This facility can be used to
non-invasively gather metrics for services utilizing HTTP. Specifying the `--prometheus-address` flag enables metrics
support. This metrics support is implemented internally as an output format, and requires an input chain that includes
the [`http`](#http) input format. A typical command for utilizing the metrics support looks like this:

```./vhs --input "tcp|http" --address 0.0.0.0:80 --capture-response --prometheus-address 0.0.0.0:8080```

This command will capture all http traffic on port 80, calculate metrics, and make them available at a `/metrics`
endpoint on port 8080 of the machine/vm/container running `vhs`.

The provided metrics include measures of request rate, error rate, and request duration, sufficient for implementing 
the [RED method](https://www.weave.works/blog/the-red-method-key-metrics-for-microservices-architecture/) of 
microservice monitoring.  Metrics are supplied on a [Prometheus](https://prometheus.io/) endpoint. Request counts are
available in a [counter vector](https://prometheus.io/docs/concepts/metric_types/#counter) labeled with HTTP method,
status code, and path. Request durations are available in a 
[summary vector](https://prometheus.io/docs/concepts/metric_types/#summary) with percentiles at 50% +/- 5%, 75% +/- 1%,
90% +/- 0.5%, 95% +/- 0.5%, 99% +/- 0.1%, 99.9% +/- 0.01%, and 99.99% +/- 0.001%. Durations are also labeled with HTTP
method, status code, and path.

## Complete Command Line Flag Reference
Command line flag               | Description
------------------------------- | -------------------------------------------------
--help, -h                      |  Show brief help for VHS.
--address string                |  Address VHS will use to capture traffic. (default "0.0.0.0:80")
--buffer-output                 |  Buffer output until the end of the flow.
--capture-response              |  Capture the responses.
--debug                         |  Emit debug logging.
--debug-http-messages           |  Emit all parsed HTTP messages as debug logs.
--debug-packets                 |  Emit all packets as debug logs.
--flow-duration duration        |  The length of the running command. (default 10s)
--gcs-bucket-name string        |  Bucket name for Google Cloud Storage
--gcs-object-name string        |  Object name for Google Cloud Storage
--http-timeout duration         |  A length of time after which an HTTP request is considered to have timed out. (default 30s)
--input string                  |  Input description.
--input-drain-duration duration |  A grace period to allow for a inputs to drain. (default 2s)
--input-file string             |  Path to an input file
--middleware string             |  A path to an executable that VHS will use as middleware.
--output strings                |  Output description.
--profile-http-address string   |  Expose profile data on this address.
--profile-path-cpu string       |  Output CPU profile to this path.
--profile-path-memory string    |  Output memory profile to this path.
--prometheus-address string     |  Address for Prometheus metrics HTTP endpoint.
--s3-compat-access-key string   |  Access key for S3-compatible storage.
--s3-compat-bucket-name string  |  Bucket name for S3-compatible storage.
--s3-compat-endpoint string     |  URL for S3-compatible storage.
--s3-compat-object-name string  |  Object name for S3-compatible storage.
--s3-compat-secret-key string   |  Secret key for S3-compatible storage.
--s3-compat-secure              |  Encrypt communication for S3-compatible storage. (default true)
--s3-compat-token string        |  Security token for S3-compatible storage.
--shutdown-duration duration    |  A grace period to allow for a clean shutdown. (default 2s)
--tcp-timeout duration          |  A length of time after which unused TCP connections are closed. (default 5m0s)

