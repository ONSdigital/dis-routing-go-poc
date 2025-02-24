FROM golang:1.24 AS build-stage

WORKDIR /app
COPY . /app/

RUN go test -v ./...
