# docker build -f cockroach.Dockerfile -t cockroacher-builder:latest .
# docker run --rm -it cockroacher-builder:latest  bash
FROM golang:1.16.4

RUN apt update \
&& apt -y install python3 autoconf cmake yacc

RUN mkdir -p $(go env GOPATH)/src/github.com/cockroachdb && cd $(go env GOPATH)/src/github.com/cockroachdb \
&& git clone https://github.com/cockroachdb/cockroach 

WORKDIR /gofuzz

# copy source files to docker
COPY goFuzz ./goFuzz
COPY sync ./sync
COPY runtime ./runtime
COPY scripts ./scripts
COPY time ./time
RUN cd goFuzz && make build

# Patch golang runtime in the container
RUN chmod +x scripts/patch-go-runtime.sh && ./scripts/patch-go-runtime.sh

