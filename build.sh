#!/usr/bin/env bash
set -euo pipefail

go version | grep -q 'go1.11' || (
    go version
    echo error: go1.11 required
    exit 1
)

go_opts="-mod=readonly"

echo [1/6] getting dependencies
go mod download > /dev/null

echo [2/6] fmt
go fmt ./...

echo [3/6] test
go test $go_opts -cover ./...

echo [4/6] lint
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

echo [5/6] build
CONF_PKG="github.com/springernature/halfpipe/config"
LDFLAGS="-X ${CONF_PKG}.VaultPrefix=springernature"
LDFLAGS="${LDFLAGS} -X ${CONF_PKG}.DocHost=docs.halfpipe.io"
LDFLAGS="${LDFLAGS} -X ${CONF_PKG}.SlackWebhook=https://hooks.slack.com/services/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"
go build $go_opts -ldflags "${LDFLAGS}" cmd/halfpipe.go

echo [6/6] e2e test
if ! e2e=$(cd e2e_test; ./test.sh 2>&1); then
    echo "${e2e}"
    exit 1
fi

echo Finished!
