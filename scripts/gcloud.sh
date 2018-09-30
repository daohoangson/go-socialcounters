#!/bin/bash

set -e
cd "$( dirname "${BASH_SOURCE[0]}" )"
cd ..

_pwd="$( pwd )"
_goPath='/root/go'
_srcPath='/root/go/src/github.com/daohoangson/go-socialcounters'
_workdir='/appengine'

docker run --rm -it \
    -e GOPATH="$_goPath" \
    -p 8000:8000 -p 8080:8080 \
    -v "$_pwd/vendor:$_goPath/src" \
    -v "$_pwd:$_srcPath" \
    -v "$_pwd/.data/empty:$_srcPath/vendor" \
    -v "$_pwd/.data/gcloud-config:/root/.config/gcloud" \
    -v "$_pwd/appengine:$_workdir" -w "$_workdir" \
    -v "$_pwd/private:$_workdir/private:ro" \
    -v "$_pwd/public:$_workdir/public:ro" \
    google/cloud-sdk bash
