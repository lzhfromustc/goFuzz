# TODO: use multi-stage build to reduce image size

FROM golang:1.16.4

RUN apt update \
&& apt -y install python3

WORKDIR /gofuzz
COPY scripts ./scripts

WORKDIR /repos
# RUN /gofuzz/scripts/clone-repos.sh

WORKDIR /gofuzz
# copy source files to docker
COPY goFuzz ./goFuzz
COPY benchmark/run.py ./benchmark/run.py
COPY sync ./sync
COPY runtime ./runtime
COPY time ./time
COPY reflect ./reflect
RUN cd goFuzz && make build

WORKDIR /playground