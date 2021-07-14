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

RUN addgroup -S gfgroup

RUN adduser -S -D gfuser gfgroup

USER gfuser

WORKDIR /gofuzz/goFuzz
RUN chmod +x ./scripts/fuzz.sh

ENTRYPOINT [ "scripts/fuzz.sh" ] 



