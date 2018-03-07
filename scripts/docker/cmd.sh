#!/bin/sh

set -e

go get -u github.com/golang/dep/cmd/dep \
	&& dep ensure \
	&& echo 'go dep ok'

go run main.go
