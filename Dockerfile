FROM golang:1.14.2

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    curl \
    hping3 \
    libpcap-dev \
    tmux

WORKDIR /go/vhs