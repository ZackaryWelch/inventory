#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VENDOR_DIR="${SCRIPT_DIR}/vendor"

mkdir -p "$VENDOR_DIR"

REDOC_REPO="Redocly/redoc"

echo "Fetching latest ReDoc release tag..."
redoc_version=$(gh release view --repo "$REDOC_REPO" --json tagName --jq '.tagName')
echo "ReDoc: ${redoc_version}"

echo "Downloading ReDoc standalone bundle..."
curl -sLo "${VENDOR_DIR}/redoc.standalone.js" \
  "https://cdn.redoc.ly/redoc/${redoc_version}/bundles/redoc.standalone.js"

echo "Verifying download..."
size=$(wc -c < "${VENDOR_DIR}/redoc.standalone.js")
if [[ "$size" -lt 1000 ]]; then
  echo "ERROR: redoc.standalone.js is only ${size} bytes, download may have failed"
  exit 1
fi
echo "  redoc.standalone.js: ${size} bytes"

echo ""
echo "Vendor assets updated successfully"
