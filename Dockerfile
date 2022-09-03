# syntax=docker/dockerfile:1

## Build

FROM golang:1.19-alpine AS build

WORKDIR /app

COPY . /app/

RUN go mod download
WORKDIR /app/cmd

RUN go build -o /orphan-finder


## Deploy
FROM alpine:latest

WORKDIR /

COPY --from=build /orphan-finder /orphan-finder
EXPOSE 8080

ENTRYPOINT ["/orphan-finder"]