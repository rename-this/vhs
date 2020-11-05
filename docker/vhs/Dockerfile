FROM golang:1.15.2-alpine3.12

RUN apk add --no-cache --virtual .build-deps \
    ca-certificates \
    g++ \
    libpcap-dev

WORKDIR /src

COPY go.sum .
COPY go.mod .

RUN go mod download

COPY . .

RUN GODEBUG=netdns=go go build \
	--ldflags '-extldflags "-static"' \
	./cmd/vhs

########################################################

FROM scratch

COPY --from=0 /src/vhs /
COPY --from=0 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/vhs"]
