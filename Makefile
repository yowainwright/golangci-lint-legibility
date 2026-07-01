GOLANGCI_LINT_CACHE ?= $(CURDIR)/.cache/golangci-lint
CUSTOM_GOFLAGS ?= -buildvcs=false

.PHONY: check custom fmt fmt-check lint test tidy-check vet

check: tidy-check fmt-check vet test lint

custom:
	GOFLAGS="$(CUSTOM_GOFLAGS)" golangci-lint custom

fmt:
	gofmt -w analyzers plugin

fmt-check:
	test -z "$$(gofmt -l analyzers plugin)"

lint: custom
	GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE) ./bin/legibility-golangci-lint run ./...

test:
	go test ./...

tidy-check:
	go mod tidy
	git diff --exit-code -- go.mod go.sum

vet:
	go vet ./...
