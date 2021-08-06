FROM golang:1.16.4

RUN apt update && apt install -y python3

WORKDIR /gofuzz

COPY goFuzz ./goFuzz
COPY sync ./sync
COPY runtime ./runtime
COPY scripts ./scripts
COPY time ./time
COPY reflect ./reflect
RUN cd goFuzz && make build
COPY benchmark ./benchmark
ENTRYPOINT [ "./benchmark/run.py" ]

