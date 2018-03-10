#!/usr/bin/env bash
set -e

echo goimports
goimports -l -w $(find . -type f -name '*.go' -not -path "./vendor/*") # ignore vendore plz..s

echo go test
go test -cover ./...

echo go vet
go vet ./...

echo dep status
if ! ds=$(dep status 2> /dev/null); then
    echo "${ds}"
    echo "Run 'dep ensure' to fix"
    exit 1
fi

#lint but ignore some of the more contentious warnings ;)
echo gometalinter
if command -v gometalinter > /dev/null; then
    gometalinter \
        --vendor \
        --disable=gocyclo \
        --exclude='should have comment' \
        --exclude='comment on exported' \
        --exclude='returns unexported type' \
        ./... || true
else
    echo "not installed. to install: go get -u github.com/alecthomas/gometalinter && gometalinter --install"
fi

echo go build

CONF_PKG="github.com/springernature/halfpipe/config"
LDFLAGS="${LDFLAGS} -X ${CONF_PKG}.VaultPrefix=springernature"
LDFLAGS="${LDFLAGS} -X ${CONF_PKG}.DocHost=docs.halfpipe.io"
LDFLAGS="${LDFLAGS} -X ${CONF_PKG}.SlackWebhook=https://hooks.slack.com/services/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"

go build -ldflags "${LDFLAGS}" cmd/halfpipe.go

echo e2e test
if ! e2e=$(cd e2e_test; ./test.sh 2>&1); then
    echo "${e2e}"
    exit 1
fi
