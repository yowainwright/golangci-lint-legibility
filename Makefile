GOLANGCI_LINT_CACHE ?= $(CURDIR)/.cache/golangci-lint
CUSTOM_GOFLAGS ?= -buildvcs=false
E2E_IMAGE ?= golangci-lint-legibility-e2e

.PHONY: check custom e2e fmt fmt-check lint test tidy-check vet

check: tidy-check fmt-check vet test e2e lint

custom:
	GOFLAGS="$(CUSTOM_GOFLAGS)" golangci-lint custom

e2e:
	docker build --file tests/e2e/Dockerfile --tag "$(E2E_IMAGE)" .
	docker run --rm "$(E2E_IMAGE)"

fmt:
	gofmt -w internal/analyzers plugin tests/e2e

fmt-check:
	test -z "$$(gofmt -l internal/analyzers plugin tests/e2e)"

lint: custom
	GOLANGCI_LINT_CACHE=$(GOLANGCI_LINT_CACHE) ./bin/legibility-golangci-lint run ./...

test:
	go test ./...

tidy-check:
	go mod tidy
	git diff --exit-code -- go.mod go.sum

vet:
	go vet ./...
