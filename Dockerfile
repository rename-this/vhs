FROM debian:10 as base

RUN apt update && \
    apt install -y  \
    libpcap0.8

########################################################

FROM gcr.io/distroless/base-debian10 as application

COPY vhs /

# Copy libpcap
COPY --from=base /usr/lib/x86_64-linux-gnu/libpcap* /usr/lib/x86_64-linux-gnu/

ENTRYPOINT ["/vhs"]

