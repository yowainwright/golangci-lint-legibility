GOLANGCI_LINT_CACHE ?= $(CURDIR)/.cache/golangci-lint

.PHONY: custom fmt fmt-check lint test vet

custom:
	golangci-lint custom

fmt:
	gofmt -w analyzers plugin

fmt-check:
	test -z "$$(gofmt -l analyzers plugin)"

lint: custom
	GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE) ./bin/legibility-golangci-lint run ./...

test:
	go test ./...

vet:
	go vet ./...
