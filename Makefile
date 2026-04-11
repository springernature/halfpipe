GO_OPTS ?=
ifdef CI
GO_OPTS = -mod=readonly
endif

default: build

build: fmt test binary e2e staticcheck dependabot schema validate-e2e

fmt:
	go fmt ./...

test:
	go test $(GO_OPTS) ./...

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

schema:
	go run ./cmd/generate-schema > schema.json

validate-e2e:
	@if ! which check-jsonschema > /dev/null 2>&1; then \
		echo "WARNING: check-jsonschema not installed, skipping schema validation of e2e tests"; \
	else \
		find .e2e -name '.halfpipe.io*' | xargs check-jsonschema --default-filetype yaml --schemafile schema.json; \
	fi

fix-e2e:
	for f in ./.e2e/*/*.actual.*; do cp "$$f" "$${f/actual/expected}"; done

coverage:
	@HALFPIPE_ENABLE_COVERAGE_TESTS=true go test $(GO_OPTS) -coverpkg=./... -coverprofile=/tmp/halfpipe-coverage.out ./... > /dev/null
	@go tool cover -func=/tmp/halfpipe-coverage.out | grep -v '^total:' | awk '\
	BEGIN { FS="\t" } \
	{ \
	  split($$1, a, ":"); path = a[1]; \
	  sub(/^github\.com\/springernature\/halfpipe\//, "", path); \
	  n = split(path, parts, "/"); \
	  pkg = ""; for (i=1; i<n; i++) pkg = pkg (i>1 ? "/" : "") parts[i]; \
	  if (pkg == "") pkg = "."; \
	  pct = $$NF; sub(/%$$/, "", pct); \
	  sum[pkg] += pct; count[pkg]++; \
	} \
	END { for (pkg in sum) printf "%-40s %.1f%%\n", pkg, sum[pkg]/count[pkg] }' | sort

.PHONY: build fmt test binary e2e staticcheck dependabot update-deps update-actions schema fix-e2e coverage validate-e2e
