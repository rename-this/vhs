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
$ ./testdata/echo.bash
```

In the other, run the following command to build `vhs`, and run it against `0.0.0.0:1111` using a dummy middleware that appends `" [[hijacked]]"` to the end of the request body for each request (note this command discards `stderr` because it can be noisy at times):

```
$ go build -o vhsout ./vhs && ./vhsout record --address 0.0.0.0:1111 --middleware ./testdata/middleware.bash | jq -R "fromjson | .body" 2> /dev/null
```

Sample output:

```
"hello, world 1594420681 [[hijacked]]"
"hello, world 1594420682 [[hijacked]]"
"hello, world 1594420683 [[hijacked]]"
"hello, world 1594420684 [[hijacked]]"
"hello, world 1594420685 [[hijacked]]"
"hello, world 1594420686 [[hijacked]]"
"hello, world 1594420687 [[hijacked]]"
"hello, world 1594420688 [[hijacked]]"
"hello, world 1594420689 [[hijacked]]"
"hello, world 1594420690 [[hijacked]]"
```