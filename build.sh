#!/usr/bin/env bash
set -euo pipefail

go version | grep -q 'go1.18' || (
    go version
    echo error: go1.18 required
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
export HALFPIPE_SKIP_COVERAGE_TESTS=true
go test $go_opts -cover ./...

echo [3/5] build
ldflags=""
if [ `git branch | grep \* | cut -d ' ' -f2` != "main" ]; then
  go build \
    $go_opts \
    -ldflags "-X github.com/springernature/halfpipe/config.CheckBranch=false" \
    cmd/halfpipe.go
else
    go build $go_opts cmd/halfpipe.go
fi

echo [4/5] e2e test
(cd e2e; ./test.sh "${1-}")

echo [5/5] staticcheck
go install honnef.co/go/tools/cmd/staticcheck@2022.1
staticcheck ./...

echo Finished!
