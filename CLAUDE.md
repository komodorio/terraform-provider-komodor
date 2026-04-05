# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Claude Code Skills & Agents

| Invoke | What it does |
|--------|-------------|
| `/release-notes` | Compares git log since last tag, groups commits by type, suggests next semver |
| `/pre-pr` | Runs `scripts/check-local.sh` and reports pass/fail per check |
| `/tf-provider-reviewer` | Reviews a resource file for CRUD completeness, schema descriptions, 404 handling, acc test coverage |

## Commands

```bash
make build          # Compile provider binary
make install        # Build and install locally for manual testing (uses OS_ARCH=darwin_amd64 by default)
make fmt            # Auto-format Go files and Terraform example files
make lint           # Run golangci-lint (5m timeout)
make generate-docs  # Regenerate docs/ from templates and schema — never edit docs/ directly
make check          # Run all CI checks locally with pass/fail summary
```

**Running tests:**
```bash
# Unit tests only
go test -count=1 -v ./komodor/... -timeout 60s

# Single test
go test -v -run TestAccKomodorRole ./komodor/... -timeout 60m

# Acceptance tests (creates real Komodor resources)
KOMODOR_API_KEY=<key> TF_ACC=1 go test -v -run TestAcc ./komodor/... -timeout 60m

# Acceptance test coverage check (no credentials needed)
go test -v -run TestAccCoverage ./komodor/...
```

## Architecture

All provider code lives in the `komodor/` package (flat structure, no sub-packages).

**Entry points:**
- `provider.go` — `Provider()` registers all resources and data sources, configures `Client` via `providerConfigure`
- `client.go` — HTTP client wrapping the Komodor API; `Client.BaseURL` drives all endpoint construction

**API endpoints:**
- `client.GetDefaultEndpoint()` → `<BaseURL>/mgmt/v1` (v1 RBAC policies)
- `client.GetV2Endpoint()` → `<BaseURL>/api/v2` (everything else: users, roles, monitors, workspaces, K8s actions, v2 policies)

Default base URL is `https://api.komodor.com`; EU region uses `https://api.eu.komodor.com`. Both are set via `api_url` provider attribute or `KOMODOR_API_URL` env var.

**Resource pattern:** Each resource has:
- `resource_komodor_<name>.go` — schema + CRUD functions
- A helper file (`roles.go`, `users.go`, `monitors.go`, etc.) — raw API calls used by the resource
- `resource_komodor_<name>_acc_test.go` — acceptance tests; must call `registerAccTest("komodor_<name>")` in `init()`

**Data sources** follow the same pattern: `datasource_komodor_<name>.go` + `_acc_test.go`.

**`common.go`** — shared helpers (e.g., reading API error responses).

**`acc_base_test.go`** — base setup for acceptance tests (cleanup of `tf-acc-` prefixed resources on test start).

**`acc_coverage_test.go`** — `TestAccCoverage` fails CI if any resource in `ResourcesMap` lacks a corresponding `_acc_test.go` file. This is enforced in Buildkite Stage 4.

## CI Pipeline (Buildkite)

Stages run sequentially with parallelism within each stage:

1. **Parallel:** fmt check, mod tidy, vet, lint, unit tests, build, docs check, examples fmt/validate, goreleaser check
2. **Acc test coverage** — `TestAccCoverage`, no credentials needed
3. **Acceptance tests** — serialized via `concurrency_group: komodor/terraform-provider/e2e`, requires `KOMODOR_API_KEY`

## Release Process

Releases are triggered by pushing a `v*` git tag. GitHub Actions runs GoReleaser, which:
- Builds multi-platform binaries (linux/darwin/windows × amd64/arm64/386/arm)
- Signs the SHA256SUMS with GPG
- Publishes to the Terraform Registry

The `VERSION` in `Makefile` is only used for local `make install` — the actual release version comes from the git tag.

## Adding a New Resource

1. Create `komodor/resource_komodor_<name>.go` with schema + CRUD
2. Create `komodor/<name>.go` (or add to an existing helper) for API calls
3. Register in `provider.go` under `ResourcesMap`
4. Create `komodor/resource_komodor_<name>_acc_test.go` with `func init() { registerAccTest("komodor_<name>") }`
5. Add examples under `examples/resources/komodor_<name>/`
6. Run `make generate-docs` and commit the updated `docs/`

CI will fail at the coverage check step if step 4 is missing.

## Key Constraints

- `docs/` is fully generated — never edit directly; run `make generate-docs` instead
- `go.sum` is generated — run `go mod tidy` instead of editing
- All acceptance test resources use the `tf-acc-` name prefix
- `komodor_policy` (v1) and `komodor_policy_v2` coexist; v1 uses `/mgmt/v1/rbac/policies`, v2 uses `/api/v2/rbac/policies`
