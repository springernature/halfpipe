#!/usr/bin/env bash
set -e

dep status

go test -cover ./...

go build -ldflags "-X main.vaultPrefix=springernature" cmd/halfpipe.go

