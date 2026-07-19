# golangci-lint-legibility

Syntax-only Go readability and comment-policy rules for `golangci-lint` module plugins.

This linter flags code that is valid but harder to scan: operator-heavy expressions, deep control flow, long equality chains, negative condition names, trivial wrappers, and similar readability issues.

## Install

`golangci-lint-legibility` is a module plugin. Build a custom `golangci-lint` binary, then run that binary with your normal `golangci-lint` config.

<!-- consumer custom-gcl config derived from go.mod module path and plugin package path -->

Create `.custom-gcl.yml` in the project that wants to use the linter:

```yaml
version: v2.12.2
name: legibility-golangci-lint
destination: ./bin
plugins:
  - module: github.com/yowainwright/golangci-lint-legibility
    import: github.com/yowainwright/golangci-lint-legibility/plugin
    version: v0.1.0
```

Replace `v0.1.0` with the release tag you want to use.

Build the custom binary:

```sh
golangci-lint custom
```

## Homebrew

<!-- Homebrew tap commands derived from Formula/golangci-lint-legibility.rb -->

Install the custom binary from this repository as a tap:

```sh
brew tap yowainwright/golangci-lint-legibility https://github.com/yowainwright/golangci-lint-legibility
brew install golangci-lint-legibility
```

## Rules

<!-- rules derived from analyzers/analyzers.go and analyzer constructors -->

Each rule has an inline do / don't diff example in [Examples](#examples).

| Code | Rule | Summary |
| --- | --- | --- |
| [`LEG001`](#leg001-max-expression-operators) | `max-expression-operators` | Limit operators inside a single expression. |
| [`LEG002`](#leg002-hoist-if-operators) | `hoist-if-operators` | Prefer named booleans before operator-heavy conditions. |
| [`LEG003`](#leg003-max-control-flow-depth) | `max-control-flow-depth` | Limit nested control-flow depth. |
| [`LEG005`](#leg005-no-quadratic-patterns) | `no-quadratic-patterns` | Flag likely quadratic nested loops. |
| [`LEG006`](#leg006-no-redundant-boolean-logic) | `no-redundant-boolean-logic` | Avoid redundant boolean comparisons. |
| [`LEG007`](#leg007-prefer-positive-condition-names) | `prefer-positive-condition-names` | Prefer positive condition names. |
| [`LEG008`](#leg008-no-trivial-wrapper-functions) | `no-trivial-wrapper-functions` | Avoid functions that only forward parameters to another call. |
| [`LEG009`](#leg009-prefer-early-return) | `prefer-early-return` | Avoid else branches after a branch already exits. |
| [`LEG010`](#leg010-prefer-guard-clauses) | `prefer-guard-clauses` | Prefer guard clauses over wrapping the main path in one large if block. |
| [`LEG011`](#leg011-max-array-chain-depth) | `max-array-chain-depth` | Limit consecutive collection-style method chains. |
| [`LEG012`](#leg012-no-computed-values) | `no-computed-values` | Prefer named values before returning computed expressions. |
| [`LEG024`](#leg024-prefer-object-lookup) | `prefer-object-lookup` | Prefer set or map lookups over long equality-or chains. |
| [`LEG025`](#leg025-require-filename-matches-dirname) | `require-filename-matches-dirname` | Opt-in. Require files in named subdirectories to match the directory name. |
| [`LEG026`](#leg026-no-mixed-filename-casing) | `no-mixed-filename-casing` | Avoid filenames that mix casing conventions. |
| [`LEG031`](#leg031-no-deep-selector-chain) | `no-deep-selector-chain` | Avoid deep selector or index chains without named intermediate values. |
| [`LEG034`](#leg034-prefer-switch-over-long-if-chain) | `prefer-switch-over-long-if-chain` | Prefer switch over long if chains that compare the same value. |
| [`LEG035`](#leg035-no-bool-literal-args) | `no-bool-literal-args` | Avoid boolean literals as call arguments. |
| [`LEG036`](#leg036-no-complex-if-init) | `no-complex-if-init` | Avoid combining an if initializer with an operator-heavy condition. |
| [`LEG037`](#leg037-no-deep-composite-literal-arg) | `no-deep-composite-literal-arg` | Avoid deeply nested composite literals as call arguments. |
| [`LEG038`](#leg038-max-function-lines) | `max-function-lines` | Limit functions to a focused line budget. |
| [`LEG039`](#leg039-no-unmatched-comments) | `no-unmatched-comments` | Policy opt-in. Reject comments without an allowed matcher or boundary identifier. |
| [`LEG040`](#leg040-no-automated-comment-attribution) | `no-automated-comment-attribution` | Reject explicit automated source signatures in comments. |
| [`LEG041`](#leg041-prefer-line-comments) | `prefer-line-comments` | Prefer `//` comments for ordinary multiline prose. |

## Configure

<!-- golangci-lint configuration derived from .golangci.yml and analyzers/settings.go -->

Add `legibility` to `.golangci.yml`:

```yaml
version: "2"

linters:
  default: standard
  enable:
    - legibility
  settings:
    custom:
      legibility:
        type: module
        description: Syntax-only Go legibility rules.
        original-url: github.com/yowainwright/golangci-lint-legibility
        settings:
          max-expression-operators: 4
          max-if-operators: 0
          max-control-flow-depth: 3
          max-array-chain-depth: 2
          max-computed-value-operators: 1
          min-object-lookup-chain-length: 3
          max-selector-chain-depth: 3
          min-switch-chain-length: 3
          max-if-init-operators: 0
          max-composite-literal-arg-depth: 1
          max-function-lines: 20
          comment-matchers:
            - '\b(ENG|OPS)-\d+\b'
          comment-prefix-identifiers:
            - HUMAN
            - LEGAL
            - SECURITY
          comment-suffix-identifiers:
            - '@owned'
          disabled-rules:
            - prefer-guard-clauses
```

Run:

```sh
./bin/legibility-golangci-lint run ./...
./bin/legibility-golangci-lint fmt ./...
```

See the `golangci-lint` [module plugin docs](https://golangci-lint.run/docs/plugins/module-plugins/) for the custom binary workflow. Module plugins require a build step because the plugin and host binary must share the same Go toolchain version and build environment.

## Recipes

<!-- workflow recipes derived from consumer binary names, golangci-lint run flags, CI checkout behavior, and comment policy settings -->

Use one committed `.golangci.yml` in every context. Change the scope and whether diagnostics block the workflow, not the rules or thresholds each context sees. Local feedback then stays representative of CI.

The examples use `./bin/legibility-golangci-lint`. Homebrew installs the same executable on `PATH`, so use `legibility-golangci-lint` without the `./bin/` prefix.

| Context | Scope | Enforcement |
| --- | --- | --- |
| Agent edit loop | Changes since `HEAD` | Blocking |
| Human exploration | New issues in the working tree | Advisory |
| Pre-commit | Changes since `HEAD` | Blocking |
| CI | Entire repository | Blocking |

### Agents

Run the linter after each coherent Go edit:

```sh
./bin/legibility-golangci-lint run --new-from-rev=HEAD ./...
```

`--new-from-rev=HEAD` includes staged, unstaged, and untracked changes. Agents should fix diagnostics introduced by their edits. They must not add ownership identifiers, broaden matchers, or add `//nolint` directives merely to pass lint.

Agents can read and preserve comments accepted by the repository policy. An unmatched comment introduced during the session should be removed or replaced by clearer code. Pre-existing diagnostics outside the task should be reported rather than suppressed or rewritten.

### Humans

Use advisory mode while exploring a change:

```sh
./bin/legibility-golangci-lint run --new --issues-exit-code=0 ./...
```

Before sharing or committing the change, remove advisory mode:

```sh
./bin/legibility-golangci-lint run --new-from-rev=HEAD ./...
```

Humans can use the configured prefix or suffix identifiers for comments that preserve intentional context. Keep regular-expression matchers narrow and reserve them for project conventions such as ticket references and tool directives. Reserve `//nolint:legibility` for an isolated false positive with an explanation.

### Pre-commit

Use the blocking form in the hook runner of your choice:

```sh
./bin/legibility-golangci-lint run --new-from-rev=HEAD ./...
```

This checks the current changes against `HEAD`. Keep the committed configuration identical to CI so a passing hook predicts a passing build.

### CI

Build the pinned custom binary, then lint the whole repository:

```sh
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
golangci-lint custom
./bin/legibility-golangci-lint run ./...
```

For a large repository with an existing baseline, retain full Git history before blocking only new PR diagnostics. Configure `actions/checkout` to keep full history:

```yaml
with:
  fetch-depth: 0
```

Fetching only `origin/main` cannot repair a shallow PR head. Refresh the base branch before running merge-base mode:

```sh
git fetch --no-tags origin +refs/heads/main:refs/remotes/origin/main
./bin/legibility-golangci-lint run --new-from-merge-base=origin/main ./...
```

Do not use `--issues-exit-code=0` in a required CI check. Cache Go modules and the `golangci-lint` cache as described in [Performance](#performance).

### Gradual adoption

Start with an advisory inventory that disables both output caps:

```sh
./bin/legibility-golangci-lint run \
  --issues-exit-code=0 \
  --max-issues-per-linter=0 \
  --max-same-issues=0 \
  ./...
```

1. Tune thresholds and exclusions for the repository without creating context-specific configs.
2. Configure comment ownership only when the repository has chosen its matcher or identifier convention.
3. Fix the remaining diagnostics, remove advisory mode, and make the full CI check required.

## Performance

Lint only changed files on PRs — 50–75% faster:

```sh
golangci-lint run --new-from-merge-base=main ./...
```

Cache `~/.cache/golangci-lint` between CI runs to avoid cold-start overhead. Example for GitHub Actions:

```yaml
- uses: actions/cache@v4
  with:
    path: ~/.cache/golangci-lint
    key: golangci-lint-${{ hashFiles('**/*.go') }}
```

## Troubleshooting

**Build step fails or binary crashes at runtime** — the plugin and `golangci-lint` must be built with the same Go toolchain. Run `go version` and confirm your toolchain matches the version in `go.mod`.

## Trust

<!-- release and provenance guarantees derived from .github/workflows/release.yml, .goreleaser.yaml, and LICENSE -->

Releases are built from pushed `v*` tags by GitHub Actions. Release tags are protected from updates and deletion, and only the repository owner can create matching release tags.

The release workflow runs `go mod tidy` drift checks, formatting, `go vet`, tests, and this linter against itself before publishing. Release jobs use pinned GitHub Actions, least-privilege permissions, and a main-branch ancestry check before publishing.

GoReleaser publishes a source archive and `checksums.txt`. GitHub releases do not publish a standalone binary; the Homebrew formula builds a custom `golangci-lint` binary from the released source.

Release artifacts are covered by GitHub artifact attestations. They use short-lived OIDC/Sigstore credentials from GitHub Actions instead of long-lived signing secrets.

Verify a release:

```sh
gh release download v0.1.0 \
  --repo yowainwright/golangci-lint-legibility \
  --pattern "golangci-lint-legibility_*_source.tar.gz" \
  --pattern "checksums.txt"

sha256sum -c checksums.txt

gh attestation verify golangci-lint-legibility_0.1.0_source.tar.gz \
  --repo yowainwright/golangci-lint-legibility
```

OpenSSF Scorecard runs weekly and reports supply-chain posture through GitHub code scanning.

## Release

<!-- release commands derived from .goreleaser.yaml and .github/workflows/release.yml -->

Create a release by signing and pushing a `v*` tag:

```sh
go mod tidy
make check

git tag -s v0.1.0
git push origin v0.1.0
```

The pushed tag triggers GoReleaser. After the workflow finishes, warm the Go module proxy:

```sh
GOPROXY=proxy.golang.org go list -m github.com/yowainwright/golangci-lint-legibility@v0.1.0
```

Update the Homebrew formula for each release by replacing the source archive URL and checksum in `Formula/golangci-lint-legibility.rb`. `brew livecheck golangci-lint-legibility` reads the latest GitHub release for update detection.

## Settings

<!-- settings derived from analyzers/settings.go -->

| Setting | Default | Description |
| --- | ---: | --- |
| `enabled-rules` | all | Only run matching rule codes or names. |
| `disabled-rules` | none | Skip matching rule codes or names. |
| `max-expression-operators` | 4 | Maximum readability operators in a single expression. |
| `max-if-operators` | 0 | Maximum boolean operators in an `if` condition. |
| `max-control-flow-depth` | 3 | Maximum nested control-flow depth. |
| `max-array-chain-depth` | 2 | Maximum consecutive collection-style method calls. |
| `max-computed-value-operators` | 1 | Maximum operators in returned or composite literal values. |
| `min-object-lookup-chain-length` | 3 | Minimum equality-or chain length before suggesting a set, map, or switch. |
| `min-dirname-match-depth` | 3 | Minimum directory depth for the opt-in filename/dirname rule. |
| `max-selector-chain-depth` | 3 | Maximum selector or index chain depth. |
| `min-switch-chain-length` | 3 | Minimum repeated comparison chain length before suggesting `switch`. |
| `max-if-init-operators` | 0 | Maximum boolean operators when an `if` also has an initializer. |
| `max-composite-literal-arg-depth` | 1 | Maximum nested composite literal depth in call arguments. |
| `max-function-lines` | 20 | Maximum source lines in a function declaration or literal; nested literals are measured independently. |
| `comment-matchers` | none | Case-insensitive Go regular expressions allowed anywhere in a comment group. |
| `comment-prefix-identifiers` | none | Literal identifiers allowed at the start of a normalized comment group. |
| `comment-suffix-identifiers` | none | Literal identifiers allowed at the end of a normalized comment group. |
| `automated-comment-identifiers` | built in | Names treated as automated sources in explicit source signatures. |
| `negative-condition-name-pattern` | built in | Regular expression for negative boolean names. |

Rule selectors accept rule codes such as `LEG009`, rule names such as `prefer-early-return`, or `all`. `require-filename-matches-dirname` is opt-in because ordinary Go packages often contain files that should not mirror the directory name. `no-unmatched-comments` activates when an allow path is configured or when the rule is explicitly selected.

## Comment policy

The comment rules form one policy stack. No ownership marker is preferred by the linter; repositories configure the convention that fits their workflow.

`no-unmatched-comments` checks each Go comment group as one unit. Consecutive `//` lines can use one prefix on the first line or one suffix on the last line. A group is accepted when any configured regular expression, prefix identifier, or suffix identifier matches. The rule activates when an allow path is configured. Explicitly selecting it with no allow path rejects every ordinary comment.

Matchers use Go regular-expression syntax, run case-insensitively against comment text without delimiters, and ignore invalid patterns. Prefix and suffix identifiers are literal, case-insensitive, and require a Go identifier boundary. Empty identifiers do not match.

The ownership rule leaves syntax and tool metadata alone:

- `//go:` compiler and tool directives, `//line`, `//extern`, `//export`, and `//nolint` directives.
- Legacy `// +build` constraints.
- Exact `// Code generated ... DO NOT EDIT.` markers before the package clause.
- Comment groups used as cgo preambles for `import "C"`.

`no-automated-comment-attribution` reports configured identifiers only when a comment explicitly assigns authorship or generation to them. Ordinary references to those technologies are unchanged. The default identifiers are `ai`, `chatgpt`, `claude`, `codex`, `copilot`, `gemini`, `gpt`, `llm`, and `openai`; set `automated-comment-identifiers: []` to disable signature matching.

`prefer-line-comments` reports multiline `/* ... */` prose because line comments are the Go norm. Single-line block comments and cgo preambles are unchanged. None of the comment rules applies an autofix.

## Examples

Removed lines are don'ts. Added lines are dos.

<!-- do/don't diff examples for every LEG rule documented in this README -->

---

### `LEG001 max-expression-operators`

#### do / don't

```diff
- return user.Active && user.Score > 10 && (user.Role == "admin" || user.Role == "owner")
+ isAdmin := user.Role == "admin"
+ isOwner := user.Role == "owner"
+ hasPrivilegedRole := isAdmin || isOwner
+ return user.Active && user.Score > 10 && hasPrivilegedRole
```

---

### `LEG002 hoist-if-operators`

#### do / don't

```diff
- if user != nil && user.Active && !user.Locked {
- 	sendInvite(user)
- }
+ canInviteUser := user != nil && user.Active && !user.Locked
+ if canInviteUser {
+ 	sendInvite(user)
+ }
```

---

### `LEG003 max-control-flow-depth`

#### do / don't

```diff
- if user != nil {
- 	for _, invite := range invites {
- 		if invite.Pending {
- 			if invite.Retries < 3 {
- 				sendInvite(invite)
- 			}
- 		}
- 	}
- }
+ if user == nil {
+ 	return
+ }
+ for _, invite := range pendingInvites(invites) {
+ 	retryInvite(invite)
+ }
```

---

### `LEG005 no-quadratic-patterns`

#### do / don't

```diff
- for _, user := range users {
- 	for _, owner := range owners {
- 		if user.ID == owner.UserID {
- 			assignOwner(user, owner)
- 		}
- 	}
- }
+ ownersByUserID := mapOwnersByUserID(owners)
+ for _, user := range users {
+ 	owner := ownersByUserID[user.ID]
+ 	assignOwner(user, owner)
+ }
```

---

### `LEG006 no-redundant-boolean-logic`

#### do / don't

```diff
- return enabled == true
+ return enabled
```

---

### `LEG007 prefer-positive-condition-names`

#### do / don't

```diff
- isNotReady := status != StatusReady
- if isNotReady {
+ isReady := status == StatusReady
+ if !isReady {
 	return nil
 }
```

---

### `LEG008 no-trivial-wrapper-functions`

#### do / don't

```diff
- func cleanName(name string) string {
- 	return strings.TrimSpace(name)
- }
-
- displayName := cleanName(input)
+ displayName := strings.TrimSpace(input)
```

---

### `LEG009 prefer-early-return`

#### do / don't

```diff
 if err != nil {
 	return err
- } else {
- 	saveUser(user)
 }
+ saveUser(user)
```

---

### `LEG010 prefer-guard-clauses`

#### do / don't

```diff
- if user.Active {
- 	validateUser(user)
- 	saveUser(user)
- }
+ if !user.Active {
+ 	return
+ }
+ validateUser(user)
+ saveUser(user)
```

---

### `LEG011 max-array-chain-depth`

#### do / don't

```diff
- hasLargeActiveOrder := orders.Filter(activeOrder).Map(orderTotal).Some(overLimit)
+ activeOrders := orders.Filter(activeOrder)
+ orderTotals := activeOrders.Map(orderTotal)
+ hasLargeActiveOrder := orderTotals.Some(overLimit)
```

---

### `LEG012 no-computed-values`

#### do / don't

```diff
- return subtotal + tax + shipping
+ taxedSubtotal := subtotal + tax
+ return taxedSubtotal + shipping
```

---

### `LEG024 prefer-object-lookup`

#### do / don't

```diff
- return role == "admin" || role == "owner" || role == "staff"
+ allowedRoles := map[string]bool{
+ 	"admin": true,
+ 	"owner": true,
+ 	"staff": true,
+ }
+ return allowedRoles[role]
```

---

### `LEG025 require-filename-matches-dirname`

Opt-in. Enable it with `enabled-rules` when the project uses directory-mirrored filenames.

#### do / don't

```diff
- internal/orders/service.go
+ internal/orders/orders_service.go
```

---

### `LEG026 no-mixed-filename-casing`

#### do / don't

```diff
- internal/API_client.go
+ internal/api_client.go
```

---

### `LEG031 no-deep-selector-chain`

#### do / don't

```diff
- return config.User.Profile.Settings.Email.Enabled
+ emailSettings := config.User.Profile.Settings.Email
+ return emailSettings.Enabled
```

---

### `LEG034 prefer-switch-over-long-if-chain`

#### do / don't

```diff
- if status == "new" {
- 	return 1
- } else if status == "active" {
- 	return 2
- } else if status == "closed" {
- 	return 3
- }
+ switch status {
+ case "new":
+ 	return 1
+ case "active":
+ 	return 2
+ case "closed":
+ 	return 3
+ }
```

---

### `LEG035 no-bool-literal-args`

#### do / don't

```diff
- createUser(user, true, false)
+ shouldSendInvite := true
+ shouldRequirePasswordReset := false
+ createUser(user, shouldSendInvite, shouldRequirePasswordReset)
```

---

### `LEG036 no-complex-if-init`

#### do / don't

```diff
- if user, ok := users[id]; ok && user.Active {
- 	save(user)
- }
+ user, ok := users[id]
+ canSaveUser := ok && user.Active
+ if canSaveUser {
+ 	save(user)
+ }
```

---

### `LEG037 no-deep-composite-literal-arg`

#### do / don't

```diff
- save(Config{HTTP: HTTPConfig{Timeout: 10}})
+ httpConfig := HTTPConfig{Timeout: 10}
+ save(Config{HTTP: httpConfig})
```

---

### `LEG038 max-function-lines`

#### do / don't

```diff
- func syncUser(user User) error {
- 	validateUser(user)
- 	normalizeUser(&user)
- 	saveUser(user)
- 	sendWelcomeEmail(user)
- 	writeAuditLog(user)
- 	refreshSearchIndex(user)
- 	return nil
- }
+ func syncUser(user User) error {
+ 	if err := prepareUser(&user); err != nil {
+ 		return err
+ 	}
+ 	return persistUser(user)
+ }
```

---

### `LEG039 no-unmatched-comments`

#### do / don't

```diff
- // Retry this request three times.
+ // HUMAN: The provider requires exactly three ordered attempts.
  for attempt := range 3 {
      retry(attempt)
  }
```

---

### `LEG040 no-automated-comment-attribution`

#### do / don't

```diff
- // <configured identifier>-generated.
+ // The provider requires retries to remain ordered.
  retry(request)
```

---

### `LEG041 prefer-line-comments`

#### do / don't

```diff
- /*
- Preserve retry order during regional failover.
- */
+ // Preserve retry order during regional failover.
  retry(request)
```

## Develop

<!-- development commands derived from Makefile -->

```sh
make tidy-check
make test
make vet
make lint
```

`make lint` builds `./bin/legibility-golangci-lint` from the included `.custom-gcl.yml` and runs it against this repo.
