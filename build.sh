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
echo golint
if command -v golint > /dev/null; then
    golint `go list ./... | grep -v /vendor/` |
        grep -v 'should have comment' |
        grep -v 'comment on exported' |
        grep -v 'returns unexported type'
else
    echo "skipping. to install: go get -u golang.org/x/lint/golint"
fi

echo go build
LD_VAULTPREFIX="-X github.com/springernature/halfpipe/config.VaultPrefix=springernature"
LD_DOCHOST="-X github.com/springernature/halfpipe/config.DocHost=docs.halfpipe.io"
LD_SLACKWEBHOOK="-X github.com/springernature/halfpipe/config.SlackWebhook=https://hooks.slack.com/services/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"
go build -ldflags "${LD_VAULTPREFIX} ${LD_DOCHOST} ${LD_SLACKWEBHOOK}" cmd/halfpipe.go

echo e2e test
if ! e2e=$(cd e2e_test; ./test.sh 2> /dev/null); then
    echo "${e2e}"
    exit 1
fi
