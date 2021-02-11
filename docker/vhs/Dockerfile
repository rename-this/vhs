FROM golang:1.15 as base

RUN go install -v -a std

RUN apt update && \
    apt install -y  libpcap-dev

WORKDIR /src

COPY go.* /src/

RUN go mod download -x

COPY . .

RUN go build ./cmd/vhs

########################################################

FROM gcr.io/distroless/base-debian10 as application

COPY --from=base /src/vhs /

# Copy libpcap
COPY --from=base /usr/include/pcap /usr/include/pcap
COPY --from=base /usr/lib/x86_64-linux-gnu/libpcap* /usr/lib/x86_64-linux-gnu/

ENTRYPOINT ["/vhs"]

