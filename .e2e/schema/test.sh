#!/usr/bin/env bash

go run ../../cmd/generate-schema > schema.actual.json

diff --ignore-blank-lines schema.actual.json schema.expected.json
