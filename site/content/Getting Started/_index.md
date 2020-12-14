---
title: "Getting Started"
linkTitle: "Getting Started"
weight: 2
description: >
    Taking VHS for a spin.
---

The quickest way to see `vhs` in action is to use the development Docker image in this repo to run a demo.

## Prerequisites
The first step is to clone the VHS repository.

```
git clone https://github.com/rename-this/vhs.git
```

You will need a working Docker installation to successfully execute the following commands. You should be able to
install Docker from your favorite package manager, or you can see the [Docker site](https://www.docker.com/get-started)
for more details.

Once you have Docker set up and working, you can build the development container (which installs some helpful tools
and mounts the source in the container) by changing directory into the repository you cloned and running:

```
$ make dev
```

## Demo Setup
Open a bash session on the container by running the following command in a terminal:

```
$ docker exect -it vhs_dev bash
```

In your new bash session, run this script to start a simple echo server and a curl request that calls the server every
second:

```
$ cd testdata && ./echo.bash
```

This will generate some local HTTP traffic for `vhs` to capture.

## Run The Demo
Open another bash session on the container in a new terminal:

```
$ docker exect -it vhs_dev bash
```

In this new session, run the following command to build `vhs` and run it to capture our local HTTP traffic.

```
$ go build ./cmd/vhs && ./vhs --input "tcp|http" --output "json|stdout" --capture-response --address 0.0.0.0:1111 --middleware ./testdata/http_middleware.bash | jq -R "fromjson | .body" 2> /dev/null
```

The output should look something like this:

```
"hello, world 1594678195 [[hijacked 0]]"
"hello, world 1594678195 [[hijacked 1]]"
"hello, world 1594678196 [[hijacked 0]]"
"hello, world 1594678196 [[hijacked 1]]"
"hello, world 1594678197 [[hijacked 0]]"
"hello, world 1594678197 [[hijacked 1]]"
"hello, world 1594678198 [[hijacked 0]]"
"hello, world 1594678198 [[hijacked 1]]"
"hello, world 1594678199 [[hijacked 0]]"
"hello, world 1594678199 [[hijacked 1]]"
"hello, world 1594678200 [[hijacked 0]]"
```

## Explanation of the Demo
We can break down the demo command we just ran to get a better feel for how `vhs` works and what it can do:

### `go build ./cmd/vhs && ./vhs`
This portion of the command builds the `vhs` from the source tree and executes the resulting binary

### `--input "tcp|http" `
This portion of the command defines input portion of the data flow for this `vhs` session. Currently, only one source
can be specified for a given `vhs` session.

In this case:
* `tcp` specifies the TCP data [source](/vhs/reference/#sources). This source will capture TCP data off the network.
* `http` specifies the HTTP [input format](/vhs/reference/#input-formats). This input format will extract HTTP requests and responses from the
  captured TCP data streams.

### `--output "json|stdout"`
This portion of the command specifies the output portion of the data flow for this `vhs` session. A `vhs` session may
have any number of outputs.

In this case:
* `json` specifies the JSON [output format](/vhs/reference/#output-formats). This output format will serialize the HTTP requests and responses into JSON.
* `stdout` specifies the standard [output sink](/vhs/reference/#sinks). This sink will print the data to the console.

### `--capture-response`
This flag tells the TCP input source to capture two-way network traffic (requests and responses) instead of one-way
(requests only).

### `--address 0.0.0.0:1111`
This flag specifies the address and port from which `vhs` will capture traffic, in the form `<IP address>:<port>`.

### `--middleware ./testdata/http_middleware.bash`
This flag specifies the middleware to run for this `vhs` session. Middleware enables users to modify or rewrite data
as it passes through `vhs` from source to sink. The middleware specified here appends
`" [[hijacked <HTTP_MESSAGE_TYPE>]]"` to the end of the http request or response body as it passes through `vhs`. More
information on middleware can be found [here](/vhs/reference/#middleware).

### `| jq -R "fromjson | .body"`
This portion of the command pipes the output of `vhs` to `jq` for filtering and pretty printing. This functionality is
external to `vhs`. You can find more information on `jq` [here](https://stedolan.github.io/jq/).

### `2> /dev/null`
Discards stderr to keep the demo output clean.

## Common Use Cases

### Capture live HTTP traffic to cloud storage
`./vhs --input "tcp|http" --output "json|gzip|gcs" --address 0.0.0.0:80 --capture-response --gcs-bucket-name <some_bucket> --gcs-object-name <some_object> --flow-duration 2m`

The above command will capture live HTTP traffic on port 80 for 2 minutes and save the captured data as gzipped JSON to
Google Cloud Storage (GCS) in the specified bucket and object.

### Generate HAR file from saved HTTP data
`./vhs --input "gcs|gzip|json --output "har|stdout" --gcs-bucket-name <some_bucket> --gcs-object-name <some_object> --flow-duration 15s > harfile.json`

The above command will "replay" saved HTTP data in gzipped JSON format from the specified GCS bucket and object and 
produce an HTTP Archive ([HAR](http://www.softwareishard.com/blog/har-12-spec/)) file called `harfile.json` on the local
filesystem.

### Provide live HTTP metrics
`./vhs --input "tcp|http" --address 0.0.0.0:80 --capture-response --prometheus-address 0.0.0.0:8888 --flow-duration 60m`

The above command will capture live HTTP traffic on port 80 and calculate 
[RED metrics](https://www.weave.works/blog/the-red-method-key-metrics-for-microservices-architecture/) for the captured
data in real time. These metrics will be available on a [Prometheus](https://prometheus.io/) endpoint at port 8888. The
command will run for 60 minutes.

