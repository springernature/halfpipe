#!/usr/bin/env sh
set -e

echo [1/5] fmt ..
go fmt ./...

echo [2/5] test ..
go test -cover ./...

echo [3/5] lint ..
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

echo [4/5] build ..
CONF_PKG="github.com/springernature/halfpipe/config"
LDFLAGS="${LDFLAGS} -X ${CONF_PKG}.VaultPrefix=springernature"
LDFLAGS="${LDFLAGS} -X ${CONF_PKG}.DocHost=docs.halfpipe.io"
LDFLAGS="${LDFLAGS} -X ${CONF_PKG}.SlackWebhook=https://hooks.slack.com/services/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"
go build -ldflags "${LDFLAGS}" cmd/halfpipe.go

echo [5/5] e2e test ..
if ! e2e=$(cd e2e_test; ./test.sh 2>&1); then
    echo "${e2e}"
    exit 1
fi

echo Finished!
