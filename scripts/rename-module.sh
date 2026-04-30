#!/usr/bin/env bash
# rename-module.sh — Replace the template module path with your own.
#
# Usage:
#   ./scripts/rename-module.sh github.com/your-org/your-project
#
# What it does:
#   1. Updates the module declaration in go.mod
#   2. Rewrites every import path in every .go file
#   3. Prints a summary of changed files
#
# Requirements: bash, sed, find (standard on Linux/macOS)

set -euo pipefail

TEMPLATE_MODULE="github.com/your-username/go-mux-backend-template"

if [[ $# -ne 1 ]]; then
  echo "Usage: $0 <new-module-path>"
  echo "Example: $0 github.com/acme/my-api"
  exit 1
fi

NEW_MODULE="$1"

if [[ "$NEW_MODULE" == "$TEMPLATE_MODULE" ]]; then
  echo "New module path is the same as the template — nothing to do."
  exit 0
fi

echo "Renaming module:"
echo "  FROM: $TEMPLATE_MODULE"
echo "  TO:   $NEW_MODULE"
echo ""

# Detect sed in-place flag (BSD/macOS vs GNU)
if sed --version 2>/dev/null | grep -q GNU; then
  SED_INPLACE="sed -i"
else
  SED_INPLACE="sed -i ''"
fi

CHANGED=0

# Update go.mod
if grep -q "$TEMPLATE_MODULE" go.mod; then
  $SED_INPLACE "s|${TEMPLATE_MODULE}|${NEW_MODULE}|g" go.mod
  echo "  ✔ go.mod"
  CHANGED=$((CHANGED + 1))
fi

# Update all .go files
while IFS= read -r -d '' file; do
  if grep -q "$TEMPLATE_MODULE" "$file"; then
    $SED_INPLACE "s|${TEMPLATE_MODULE}|${NEW_MODULE}|g" "$file"
    echo "  ✔ $file"
    CHANGED=$((CHANGED + 1))
  fi
done < <(find . -name "*.go" -not -path "./vendor/*" -print0)

echo ""
echo "Done. $CHANGED file(s) updated."
echo ""
echo "Next steps:"
echo "  go mod tidy"
echo "  make sqlc-gen   # regenerate the repository package"
echo "  make build      # verify everything compiles"
