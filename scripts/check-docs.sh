#!/usr/bin/env bash
set -euo pipefail

echo "Regenerating provider docs..."
make generate-docs

if ! git diff --exit-code docs/; then
  echo ""
  echo "ERROR: docs/ is out of date."
  echo "Run 'make generate-docs' locally and commit the result."
  exit 1
fi

echo "docs/ is up to date."
