FROM golang:1.11.11-alpine3.10

RUN apk add --update build-base git

VOLUME ["/usr/local/go/src"]
