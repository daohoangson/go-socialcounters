#!/bin/bash

set -e
cd "$( dirname "${BASH_SOURCE[0]}" )/.."
_pwd="$( pwd )"

exec docker-compose run --rm \
  --publish 8080:8080 \
  --volume "$_pwd:/src" \
  --volume "$_pwd/.data/go/pkg/mod:/go/pkg/mod" \
  golang sh
