#!/usr/bin/env bash
set -euo pipefail

[[ -d /var/halfpipe/shared-cache ]] && export GOPATH="/var/halfpipe/shared-cache"

go version | grep -q 'go1.15' || (
    go version
    echo error: go1.15 required
    exit 1
)

go_opts=""
if [ "${1-}" = "ci" ]; then
    echo CI
    go_opts="-mod=readonly"
fi

echo [1/5] fmt
go fmt ./...

echo [2/5] test
go test $go_opts -cover ./...

echo [3/5] build
ldflags=""
if [ `git branch | grep \* | cut -d ' ' -f2` != "master" ]; then
  go build \
    $go_opts \
    -ldflags "-X github.com/springernature/halfpipe/config.CheckBranch=false" \
    cmd/halfpipe.go
else
    go build $go_opts cmd/halfpipe.go
fi

echo [4/5] e2e test
(cd e2e; ./test.sh "${1-}")

echo [5/5] lint
if command -v golint > /dev/null; then
    golint ./... |
        grep -v 'should have comment or be unexported' |
        grep -v 'returns unexported type' \
    || true
else
    echo "golint not installed. to install: GO111MODULE=off go get -u golang.org/x/lint/golint"
fi

echo Finished!
