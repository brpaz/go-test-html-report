#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR=$(dirname "$(dirname "$(realpath "$0")")")
cd "$ROOT_DIR"

read -r -p "Enter the package name (e.g., github.com/username/repo): " PACKAGE_NAME

if [[ -z "$PACKAGE_NAME" ]]; then
  echo "Package name cannot be empty"
  exit 1
fi

# Replace package name in all files
grep -rl 'github.com/brpaz/go-test-html-report' . | xargs sed -i "s|github.com/brpaz/go-test-html-report|${PACKAGE_NAME}|g"
