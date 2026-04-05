---
name: pre-pr
description: Run all local quality checks (fmt, vet, lint, unit tests, docs, terraform validate) before opening a PR
user-invocable: true
---

Run: bash scripts/check-local.sh

Report pass/fail for each check. If anything fails, show the relevant output and suggest the exact fix command.
