FROM golang:1.14.2

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    hping3 \
    libpcap-dev \
    tmux

WORKDIR /src

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build ./cmd/vhs

ENTRYPOINT ["tmux"]