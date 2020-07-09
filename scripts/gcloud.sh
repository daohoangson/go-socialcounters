#!/bin/bash

set -e
cd "$( dirname "${BASH_SOURCE[0]}" )/.."
_pwd="$( pwd )"

docker build \
  --file ./appengine/gcloud.Dockerfile \
  --tag tmp \
  .

exec docker run --rm -it \
  --publish 8000:8000 --publish 8080:8080 \
  --volume "$_pwd:/src" \
  --volume "$_pwd/.data/gcloud-config:/root/.config/gcloud" \
  tmp
