#!/usr/bin/env bash

go test -cover ./...

go build cmd/halfpipe.go

