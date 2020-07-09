FROM golang:1.14.2

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    curl \
    hping3 \
    jq \
    libpcap-dev \
    tcpdump \
    tmux

RUN go get github.com/go-delve/delve/cmd/dlv

WORKDIR /go/vhs