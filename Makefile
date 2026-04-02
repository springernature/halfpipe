default: build

build:
	./build.sh

update-deps:
	go get -t -u ./... && go mod tidy

update-actions:
	go run ./cmd/update-actions

fix-e2e:
	@for d in ./e2e/actions/*/; do \
		[ -f "$$d/workflowActual.yml" ] && cp $$d/workflowActual.yml $$d/workflowExpected.yml; \
	done; true
	@for d in ./e2e/concourse/*/; do \
		[ -f "$$d/pipelineActual.yml" ] && cp $$d/pipelineActual.yml $$d/pipelineExpected.yml; \
	done; true

.PHONY: build update-deps update-actions fix-e2e


