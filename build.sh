#!/usr/bin/env bash
set -e

go test -cover ./...

go build -ldflags "-X main.vaultPrefix=springernature" cmd/halfpipe.go

