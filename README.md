# vhs

Record and replay HTTP traffic

# Development

The easiest way to run this is to use the docker image included in the repo.

To build the image (which installs some helpful tools and mounts the source in the container):

```
$ make dev
```

Open two bash sessions on the container. 

In one, run this script to boot a simple echo server and a curl request that calls the server every second:

```
$ cd testdata && ./echo.bash
```

In the other, run the following command to build `vhs`, and run it against `0.0.0.0:1111` using a dummy middleware that appends `" [[hijacked <HTTP_MESSAGE_TYPE>]]"` to the end of the request or response body (note this command discards `stderr` because it can be noisy at times):

```
$ go build ./cmd/vhs && ./vhs --input "tcp|http" --output "json|stdout" --capture-response --address 0.0.0.0:1111 --middleware ./testdata/http_middleware.bash | jq -R "fromjson | .body" 2> /dev/null
```

Sample output:

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

## GCS

To write output to GCS, first create a container that you have the rights to write to.

Create a JSON keyfile:

```
$ gcloud auth application-default login
```

Note the location of this file, mine is `$HOME/.config/gcloud/application_default_credentials.json`.

Running `make dev` will assume this location and mount the dir from that file in the docker container, keep this in mind when you boot it up.

When running `vhs`, you pass the bucket name like this:

```
$ go build ./cmd/vhs && ./vhs --input "tcp|http" --output "json|gzip|gcs" --capture-response --address 0.0.0.0:1111 --gcs-bucket-name vhs-test-andrewhare
```
