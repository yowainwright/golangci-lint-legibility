# Contributing

## Development

<!-- local development commands derived from go.mod, .custom-gcl.yml, .golangci.yml, and Makefile -->

Install Go and `golangci-lint` v2, then run:

```sh
go mod download
make test
make vet
make lint
```

`make lint` builds `./bin/legibility-golangci-lint` from `.custom-gcl.yml` and runs the custom linter against this repo.

## Code Style

Keep analyzer logic syntax-only unless a rule explicitly needs type information. Prefer small helpers, early returns, and named intermediate values for complex conditions.

## Pull Requests

Open focused pull requests with tests for new or changed rules. Include a short before-and-after example when changing diagnostics or settings.
