# TODO: use multi-stage build to reduce image size

FROM golang:1.16.4

RUN apt update \
&& apt -y install python3

WORKDIR /gofuzz

# copy source files to docker
COPY . .

RUN cd goFuzz \
&& make build

# Patch golang runtime in the container
RUN chmod +x scripts/patch-go-runtime.sh \
&& ./scripts/patch-go-runtime.sh


WORKDIR /gofuzz/goFuzz

RUN groupadd gfgroup
RUN useradd -r -u 1001 -g gfgroup gfuser
RUN chown gfuser:gfgroup ./scripts/fuzz.sh && chmod +x ./scripts/fuzz.sh

USER gfuser

ENTRYPOINT [ "scripts/fuzz.sh" ] 



