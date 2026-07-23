#!/usr/bin/env bash
#
# regen-docs.sh — regenerate the provider's Terraform Registry docs (docs/).
#
# Runs `go generate ./...` from the repo root: `terraform fmt` on the examples,
# then `tfplugindocs generate` to re-render every docs/**/*.md from the live
# provider schema + templates/ + examples/. Then runs `tfplugindocs validate`
# (publishability: frontmatter, size, schema<->doc parity).
#
# Docs are 100% generated — never hand-edit docs/*.md. Change the schema
# (upstream), templates/, or examples/ and re-run this.
#
# Usage:
#   regen-docs.sh
#
# Env:
#   CHECK=1   also fail on drift after generating — reproduces the docsci CI gate
#             exactly (use to confirm the committed docs are fresh; NOT while
#             authoring, where new/changed docs are expected).
#
set -euo pipefail

WT="$(git rev-parse --show-toplevel)"
cd "$WT"

command -v terraform >/dev/null 2>&1 || {
  echo "terraform not on PATH — go generate needs it to fmt the examples" >&2
  exit 1
}

echo ">> go generate ./...   (terraform fmt examples + tfplugindocs generate)"
go generate ./...

if [ -n "${CHECK:-}" ]; then
  echo ">> drift check (git diff --exit-code)"
  git diff --compact-summary --exit-code || {
    echo >&2
    echo "docs are stale — run without CHECK, then commit the regenerated docs/ + examples/." >&2
    exit 1
  }
else
  echo ">> changed files (docs/ + examples/):"
  git diff --compact-summary -- docs/ examples/ || true
fi

echo ">> tfplugindocs validate"
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs validate -provider-name signoz

echo ">> docs regenerated and validated."
