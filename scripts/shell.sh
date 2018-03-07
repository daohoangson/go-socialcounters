#!/bin/bash

set -e
cd "$( dirname "${BASH_SOURCE[0]}" )"
cd ..

docker-compose run --rm -p 8080:8080 golang sh
