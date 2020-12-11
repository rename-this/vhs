FROM golang:1.15.6-alpine3.12 as base

RUN apk add --no-cache --virtual .build-deps \
	bash \
	ca-certificates \
	g++ \
	jq \
	libpcap-dev

RUN go install -v -a std

WORKDIR /src

COPY go.* /src/

RUN go mod download -x

COPY . .

########################################################

FROM base as builder

RUN GODEBUG=netdns=go \
    go build \
    -ldflags "-s -extldflags=-static" \
    ./cmd/vhs

########################################################

FROM scratch as application

COPY --from=builder /src/vhs /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/vhs"]

