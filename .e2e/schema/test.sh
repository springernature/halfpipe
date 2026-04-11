#!/usr/bin/env bash

go run ../../cmd/generate-schema > schema.actual.json

echo "Schema"
diff --ignore-blank-lines schema.actual.json schema.expected.json
