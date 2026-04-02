#!/usr/bin/env bash
set -euo pipefail

echo "Snapshotting current docs/..."
TMPDIR=$(mktemp -d)
cp -r docs/ "${TMPDIR}/docs-before"

echo "Regenerating provider docs..."
make generate-docs

echo "Comparing docs/..."
if ! diff -rq "${TMPDIR}/docs-before" docs/ > /dev/null 2>&1; then
  diff -r "${TMPDIR}/docs-before" docs/ || true
  echo ""
  echo "ERROR: docs/ is out of date."
  echo "Run 'make generate-docs' locally and commit the result."
  rm -rf "${TMPDIR}"
  exit 1
fi

rm -rf "${TMPDIR}"
echo "docs/ is up to date."
