#!/usr/bin/env sh
set -eu

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

git config core.hooksPath scripts/hooks
chmod +x scripts/hooks/pre-commit scripts/hooks/post-merge

printf '%s\n' "Configured git hooks from scripts/hooks."
printf '%s\n' "pre-commit and post-merge now run: make check"
