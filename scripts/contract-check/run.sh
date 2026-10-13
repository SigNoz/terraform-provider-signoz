#!/usr/bin/env bash
# Runs scripts/contract-check/main.go inside a throwaway module that `replace`s
# both the provider (this repo) and the SigNoz server, so the provider's
# go.mod never depends on the server. Usage: run.sh <path-to-signoz-checkout>.
set -euo pipefail

repo="$(cd "$(dirname "$0")/../.." && pwd)"
signoz="${1:?usage: run.sh <path-to-signoz-checkout>}"
signoz="$(cd "$signoz" && pwd)"

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

cat > "$tmp/go.mod" <<EOF
module github.com/SigNoz/terraform-provider-signoz/internal/contractcheck

go 1.25.8

require (
	github.com/SigNoz/signoz v0.0.0
	github.com/SigNoz/terraform-provider-signoz v0.0.0
)

replace github.com/SigNoz/signoz => $signoz

replace github.com/SigNoz/terraform-provider-signoz => $repo
EOF

# Drop the ignore build tag so the throwaway module can build the file.
sed '/^\/\/go:build ignore$/d' "$repo/scripts/contract-check/main.go" > "$tmp/main.go"

cd "$tmp"
GOWORK=off GOTOOLCHAIN=go1.25.8 GOFLAGS=-mod=mod go mod tidy
GOWORK=off GOTOOLCHAIN=go1.25.8 go run .
