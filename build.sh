#!/usr/bin/env bash

go test -cover ./...

go build -ldflags "-X main.vaultPrefix=springernature" cmd/halfpipe.go

