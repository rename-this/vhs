FROM golang:1.14.2

RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] http://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
RUN curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key --keyring /usr/share/keyrings/cloud.google.gpg add -

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    curl \
    google-cloud-sdk \
    hping3 \
    jq \
    libpcap-dev \
    tcpdump \
    telnet \
    tmux

WORKDIR /go/vhs