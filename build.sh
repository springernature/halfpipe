#!/usr/bin/env bash
set -euo pipefail

go version | grep -q 'go1.11' || (
    go version
    echo error: go1.11 required
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

echo [3/5] lint
# gometalinter not happy with 1.11
# if command -v gometalinter > /dev/null; then
#     gometalinter --fast \
#         --disable=gocyclo \
#         --exclude='should have comment' \
#         --exclude='comment on exported' \
#         --exclude='returns unexported type' \
#         ./... || true
# else
#     echo "not installed. to install: go get -u github.com/alecthomas/gometalinter && gometalinter --install"
# fi

echo [4/5] build
CONF_PKG="github.com/springernature/halfpipe/config"
LDFLAGS="-X ${CONF_PKG}.DocHost=docs.halfpipe.io"
LDFLAGS="${LDFLAGS} -X ${CONF_PKG}.SlackWebhook=https://hooks.slack.com/services/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"
go build $go_opts -ldflags "${LDFLAGS}" cmd/halfpipe.go

echo [5/5] e2e test
cd e2e; ./test.sh "${1-}"

echo Finished!
