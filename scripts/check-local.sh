#!/usr/bin/env bash
# Runs every CI check locally and prints a summary of what passed and what failed.
# Skips checks whose required tool is not installed.
# Usage: ./scripts/check-local.sh
set -uo pipefail

# Required for all go commands — override whatever the shell has.
export GO111MODULE=on

# ── Colours ────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

# ── Result tracking ────────────────────────────────────────────────────────
declare -a RESULTS   # "PASS|SKIP|FAIL  <label>"
FAILED=0

pass()  { RESULTS+=("${GREEN}PASS${RESET}  $1"); }
fail()  { RESULTS+=("${RED}FAIL${RESET}  $1"); FAILED=1; }
skip()  { RESULTS+=("${YELLOW}SKIP${RESET}  $1"); }

# ── Helper: run a check ────────────────────────────────────────────────────
# run_check <label> <required-binary> <command...>
run_check() {
  local label="$1"
  local required="$2"
  shift 2

  echo -e "\n${CYAN}${BOLD}── ${label}${RESET}"

  if [ -n "${required}" ] && ! command -v "${required}" &>/dev/null; then
    echo "  skipped: '${required}' not found in PATH"
    skip "${label}"
    return
  fi

  if "$@" 2>&1; then
    pass "${label}"
  else
    fail "${label}"
  fi
}

# ── Checks ─────────────────────────────────────────────────────────────────

run_check "fmt check" "gofmt" bash -c '
  unformatted=$(gofmt -l .)
  if [ -n "$unformatted" ]; then
    echo "Files need formatting (run: gofmt -w .):"
    echo "$unformatted"
    exit 1
  fi
'

run_check "mod tidy" "go" bash -c '
  go mod tidy
  if ! git diff --exit-code go.mod go.sum; then
    echo "go.mod/go.sum are out of sync — run: go mod tidy"
    exit 1
  fi
'

run_check "vet" "go" go vet ./...

run_check "lint" "golangci-lint" make lint


run_check "unit tests" "go" go test -race -count=1 -v ./komodor/... -timeout 60s

run_check "build" "go" go build -v ./...

run_check "docs check" "go" bash scripts/check-docs.sh

run_check "examples fmt" "terraform" terraform fmt -check -recursive examples/

run_check "examples validate" "go" bash scripts/validate-examples.sh

run_check "goreleaser check" "goreleaser" goreleaser check

run_check "acc test coverage" "go" go test -v -run TestAccCoverage ./komodor/...

# Acceptance tests: only run when KOMODOR_API_KEY is set
echo -e "\n${CYAN}${BOLD}── acceptance tests${RESET}"
if [ -z "${KOMODOR_API_KEY:-}" ]; then
  echo "  skipped: set KOMODOR_API_KEY to run"
  skip "acceptance tests"
else
  if TF_ACC=1 go test -v -run TestAcc ./komodor/... -timeout 60m 2>&1; then
    pass "acceptance tests"
  else
    fail "acceptance tests"
    FAILED=1
  fi
fi

# ── Summary ────────────────────────────────────────────────────────────────
echo -e "\n${BOLD}════════════════════════════════${RESET}"
echo -e "${BOLD}Results${RESET}"
echo -e "${BOLD}════════════════════════════════${RESET}"
for r in "${RESULTS[@]}"; do
  echo -e "  ${r}"
done
echo ""

if [ "${FAILED}" -ne 0 ]; then
  echo -e "${RED}${BOLD}Some checks failed. See output above for details.${RESET}"
  exit 1
else
  echo -e "${GREEN}${BOLD}All checks passed.${RESET}"
fi
