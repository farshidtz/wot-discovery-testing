FROM golang:1.15-alpine

COPY . /home

WORKDIR /home

ENV CGO_ENABLED=0

ENTRYPOINT ["go", "test", "-v"]