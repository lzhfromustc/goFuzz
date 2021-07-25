# TODO: use multi-stage build to reduce image size
# TODO: extend from built goFuzz base to remove duplicated logic
FROM golang:1.16.4

RUN apt update \
&& apt -y install python3

WORKDIR /gofuzz

# copy source files to docker
COPY goFuzz ./goFuzz
COPY sync ./sync
COPY runtime ./runtime
COPY scripts ./scripts
RUN cd goFuzz && make build

# Patch golang runtime in the container
RUN chmod +x scripts/patch-go-runtime.sh \
&& ./scripts/patch-go-runtime.sh

WORKDIR /gofuzz/goFuzz

# RUN groupadd gfgroup
# RUN useradd --create-home -r -u 1001 -g gfgroup gfuser
# RUN chown gfuser:gfgroup ./scripts/fuzz.sh && chmod +x ./scripts/fuzz.sh
# USER gfuser

RUN chmod +x ./scripts/fuzz.sh
ENTRYPOINT [ "scripts/fuzz.sh" ] 