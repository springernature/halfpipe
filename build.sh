#!/usr/bin/env bash
set -euo pipefail

go version | grep -q 'go1.21' || (
    go version
    echo error: go1.21 required
    exit 1
)

go_opts=""
if [ "${1-}" = "ci" ]; then
    echo CI
    go_opts="-mod=readonly"
fi


# https://github.com/golang/go/wiki/LoopvarExperiment
export GOEXPERIMENT=loopvar

echo [1/6] fmt
go fmt ./...

echo [2/6] test
export HALFPIPE_SKIP_COVERAGE_TESTS=true
go test $go_opts -cover ./...

echo [3/6] build
if [ "$(git branch | grep "\*" | cut -d ' ' -f2)" != "main" ]; then
  go build \
    $go_opts \
    -ldflags "-X github.com/springernature/halfpipe/config.CheckBranch=false" \
    cmd/halfpipe.go
else
    go build $go_opts cmd/halfpipe.go
fi

echo [4/6] e2e test
(cd e2e; ./test.sh)

echo [5/6] staticcheck
staticcheck -version 2> /dev/null | grep -Fq '2023.1.6' || go install honnef.co/go/tools/cmd/staticcheck@2023.1.6
staticcheck ./...

echo [6/6] update dependabot workflow
./halfpipe -q -i dependabot.halfpipe.io

echo Finished!
