#!/usr/bin/env bash
#
# run-pipeline.sh — run the ENTIRE skaff codegen pipeline deterministically.
#
# Regenerates every resource declared in skaff.yml from the given OpenAPI spec,
# in the required order (convertors LAST, with --callers-dir; --convtypes-import
# on both flex and convertors). skaff is run through primus.
#
# It does NOT edit skaff.yml, register resources, regenerate docs, or open a PR —
# the resource-creator / spec-syncer skills own those steps around this one.
#
# Usage:
#   run-pipeline.sh <openapi-spec-path>
#
# Env:
#   DRY_RUN=1        print each skaff invocation instead of running it
#   SKAFF_VERSION=vX.Y.Z   pin the skaff release primus downloads (default: primus default)
#
set -euo pipefail

SPEC="${1:-}"
if [ -z "$SPEC" ] || [ ! -f "$SPEC" ]; then
  echo "usage: $0 <openapi-spec-path>   (spec file not found)" >&2
  exit 2
fi
SPEC="$(cd "$(dirname "$SPEC")" && pwd)/$(basename "$SPEC")"   # absolutize

WT="$(git rev-parse --show-toplevel)"
MOD="$(awk '/^module /{print $2; exit}' "$WT/go.mod")"

# Resolve primus (PRIMUS_HOME, else a local ./.primus clone).
PRIMUS_REL="src/make/main.mk"
if [ -n "${PRIMUS_HOME:-}" ] && [ -f "$PRIMUS_HOME/$PRIMUS_REL" ]; then
  PRIMUS_MK="$PRIMUS_HOME/$PRIMUS_REL"
elif [ -f "$WT/.primus/$PRIMUS_REL" ]; then
  PRIMUS_MK="$WT/.primus/$PRIMUS_REL"
else
  echo "primus not set up — run the primus-setter skill (or export PRIMUS_HOME)" >&2
  exit 1
fi

# Run one skaff subcommand through primus.
skaff() {
  echo ">> skaff $1"

  local -a mkargs=(-f "$PRIMUS_MK" skaff "SKAFF_ARGS=$*")
  [ -n "${SKAFF_VERSION:-}" ] && mkargs+=("SKAFF_VERSION=$SKAFF_VERSION")

  if [ -n "${DRY_RUN:-}" ]; then
    printf '   make'; printf ' %q' "${mkargs[@]}"; printf '\n'
  else
    make "${mkargs[@]}"
  fi
}

echo ">> skaff pipeline  spec=$SPEC  repo=$WT  module=$MOD"

skaff types --openapi "$SPEC" --config "$WT/skaff.yml" --output "$WT/internal/customtypes"

skaff schemas --openapi "$SPEC" --config "$WT/skaff.yml" \
  --types-dir "$WT/internal/customtypes" --output "$WT/internal/schemas" \
  --package schemas --module "$MOD"

skaff apitypes --openapi "$SPEC" --config "$WT/skaff.yml" \
  --output "$WT/internal/apitypes" --package apitypes

skaff client --openapi "$SPEC" --config "$WT/skaff.yml" \
  --output "$WT/internal/apiclients/zz_generated_client.go" --package apiclients \
  --apitypes-import "$MOD/internal/apitypes"

skaff flex --openapi "$SPEC" --config "$WT/skaff.yml" \
  --schemas-dir "$WT/internal/schemas" --apitypes-dir "$WT/internal/apitypes" \
  --types-dir "$WT/internal/customtypes" --output "$WT/internal/convertors" --package conv \
  --apitypes-import "$MOD/internal/apitypes" \
  --schemas-import "$MOD/internal/schemas" \
  --customtypes-import "$MOD/internal/customtypes" \
  --convtypes-import "$MOD/internal/convtypes"

skaff services --openapi "$SPEC" --config "$WT/skaff.yml" \
  --output "$WT/internal/services" --package services \
  --schemas-dir "$WT/internal/schemas" --convertors-dir "$WT/internal/convertors" \
  --apiclients-import "$MOD/internal/apiclients" \
  --apitypes-import "$MOD/internal/apitypes" \
  --schemas-import "$MOD/internal/schemas" \
  --convertors-import "$MOD/internal/convertors"

# convertors LAST, with --callers-dir: its reachability roots include the flex +
# services output, so it must run after both (see docs/convertors.md).
skaff convertors \
  --types-dir "$WT/internal/customtypes" --apitypes-dir "$WT/internal/apitypes" \
  --output "$WT/internal/convertors" --package conv \
  --apitypes-import "$MOD/internal/apitypes" \
  --customtypes-import "$MOD/internal/customtypes" \
  --convtypes-import "$MOD/internal/convtypes" \
  --callers-dir "$WT/internal/services"

echo ">> pipeline complete — review 'skipped N' from flex/services (must be 0), then build/lint"
