#!/usr/bin/env bash
set -euo pipefail

echo "Snapshotting current docs/..."
WORK_DIR=$(mktemp -d)
trap 'rm -rf "${WORK_DIR}"' EXIT
cp -r docs/ "${WORK_DIR}/docs-before"

echo "Regenerating provider docs..."
make generate-docs

echo "Comparing docs/..."
if ! diff -rq "${WORK_DIR}/docs-before" docs/ > /dev/null 2>&1; then
  diff -r "${WORK_DIR}/docs-before" docs/ || true
  echo ""
  echo "ERROR: docs/ is out of date."
  echo "Run 'make generate-docs' locally and commit the result."
  exit 1
fi

echo "docs/ is up to date."
