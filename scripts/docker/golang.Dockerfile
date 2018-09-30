FROM golang:1.11.0-alpine3.7

RUN apk add --update git

VOLUME ["/usr/local/go/src"]
