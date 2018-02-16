#!/usr/bin/env bash
set -ex

dep status

go test -cover ./...

go build -ldflags "-X main.vaultPrefix=springernature" cmd/halfpipe.go

