FROM golang:1.15 as builder

RUN go install -v -a std

RUN apt update && \
    apt install -y \
    ca-certificates \
    g++ \
    jq \
    libpcap-dev

WORKDIR /src

COPY go.* /src/

RUN go mod download

COPY . .

RUN GODEBUG=netdns=go \
    go build \
    -ldflags "-s" \
    ./cmd/vhs

########################################################

FROM scratch as application

COPY --from=builder /src/vhs /
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/vhs"]

########################################################

FROM builder as test

RUN make test

