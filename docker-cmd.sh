#!/bin/sh

go get github.com/tools/godep \
	&& godep restore \
	&& echo 'godep ok'

go run main.go