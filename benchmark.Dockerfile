FROM golang:1.16.4

RUN apt update && apt install -y python3

WORKDIR /gofuzz

# copy source files to docker
COPY goFuzz ./goFuzz
COPY sync ./sync
COPY runtime ./runtime
COPY scripts ./scripts
COPY time ./time
COPY reflect ./reflect
RUN cd goFuzz && make build

# Patch golang runtime in the container
RUN chmod +x scripts/patch-go-runtime.sh \
&& ./scripts/patch-go-runtime.sh


COPY benchmark ./benchmark

ENTRYPOINT [ "./benchmark/run.py" ]

