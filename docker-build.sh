#!/usr/bin/env bash

docker run -it \
  -v "$(go env GOMODCACHE)":/gomodcache \
  -v "$PWD":/halfpipe \
  -w /halfpipe \
  -e GOMODCACHE=/gomodcache \
  golang:1.23 \
  ${1:-./build.sh}
