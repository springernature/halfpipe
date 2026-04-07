GO_OPTS ?=
ifdef CI
GO_OPTS = -mod=readonly
endif

default: build

build: fmt test binary e2e staticcheck dependabot

fmt:
	go fmt ./...

test:
	go test $(GO_OPTS) -cover ./...

binary:
	go build $(GO_OPTS) -o halfpipe cmd/halfpipe.go

e2e: binary
	.e2e/test.sh

staticcheck:
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...

dependabot: binary
	./halfpipe -q -i dependabot.halfpipe.io.yml

update-deps:
	go get -t -u ./... && go mod tidy

update-actions:
	go run ./cmd/update-actions

fix-e2e:
	for f in ./.e2e/*/*actual*.yml; do cp "$$f" "$${f/actual/expected}"; done

.PHONY: build fmt test binary e2e staticcheck dependabot update-deps update-actions fix-e2e
