#!/usr/bin/env bash
set -euo pipefail

go version | grep -q 'go1.26' || (
    go version
    echo error: go1.26 required
    exit 1
)

go_opts=""
if [ "${1-}" = "ci" ]; then
    echo CI
    go_opts="-mod=readonly"
fi

echo [1/6] fmt
go fmt ./...

echo [2/6] test
go test $go_opts -cover ./...

echo [3/6] build
go build $go_opts cmd/halfpipe.go

echo [4/6] e2e test
(cd e2e; ./test.sh)

echo [5/6] staticcheck
go run honnef.co/go/tools/cmd/staticcheck@latest ./...

echo [6/6] update dependabot workflow
./halfpipe -q -i dependabot.halfpipe.io

echo Finished!
