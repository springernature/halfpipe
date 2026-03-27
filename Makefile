default: build

build:
	./build.sh

update-deps:
	go get -t -u ./... && go mod tidy

update-actions:
	go run ./cmd/update-actions

.PHONY: build update-deps update-actions
