FROM golang:1.16.4

RUN apt update && apt install -y python3

WORKDIR /playground

CMD [ "bash" ]

