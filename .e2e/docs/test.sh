#!/usr/bin/env bash

go run ../../cmd/generate-docs > docs.actual.md

diff --ignore-blank-lines docs.actual.md docs.expected.md
