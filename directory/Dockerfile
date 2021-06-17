FROM golang:1.15-alpine

COPY . /home

WORKDIR /home

ENV CGO_ENABLED=0

# download dependencies
RUN go get .../.

ENTRYPOINT ["go", "test"]