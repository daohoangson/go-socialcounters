FROM golang:1.10.0-alpine3.7

RUN apk add --update git

VOLUME ["/usr/local/go/src"]
