#!/bin/bash
# Setup git hooks for this repository

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$REPO_ROOT"

echo "Setting up git hooks..."
git config core.hooksPath .githooks

echo "✅ Git hooks configured!"
echo ""
echo "Pre-commit hook will run:"
echo "  - go build ./..."
echo "  - go test ./..."
echo "  - golangci-lint run"
echo ""
echo "Make sure golangci-lint v2 is installed:"
echo "  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"
