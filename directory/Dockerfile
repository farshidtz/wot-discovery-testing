FROM golang:1.17-alpine

COPY . /home

WORKDIR /home

ENV CGO_ENABLED=0

# download dependencies
RUN go mod download

ENTRYPOINT ["go", "test"]