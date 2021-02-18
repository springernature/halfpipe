#!/usr/bin/env bash

mkdir -p ~/.halfpipe-shared-cache

docker run -it \
  -v ~/.halfpipe-shared-cache:/var/halfpipe/shared-cache \
  -v "$PWD":/halfpipe \
  -w /halfpipe \
  golang:1.16-buster \
  ./build.sh
