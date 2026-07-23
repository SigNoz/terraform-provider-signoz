#!/usr/bin/env bash
#
# setup.sh — set up SigNoz primus for this repo: clone it and export PRIMUS_HOME.
#
# Checks first: if primus is already set up (PRIMUS_HOME valid), does NOTHING.
# Persists PRIMUS_HOME to your shell profile so it survives new shells.
#
# Usage:
#   setup.sh [DEST]     DEST = where to clone primus.
#                       Empty → ".primus" at the repo root (gitignored).
#
set -euo pipefail

MAIN_MK_REL="src/make/main.mk"

repo_root() { git rev-parse --show-toplevel; }

# 0. Already set up? Never re-clone or touch the profile.
if [ -n "${PRIMUS_HOME:-}" ] && [ -f "$PRIMUS_HOME/$MAIN_MK_REL" ]; then
  echo "primus already set up: PRIMUS_HOME=$PRIMUS_HOME — nothing to do."
  exit 0
fi

# 1. Destination: the given path, else the default .primus in the repo.
dest="${1:-}"
if [ -z "$dest" ]; then
  dest="$(repo_root)/.primus"
fi

# Absolutize (the parent must exist to resolve it).
mkdir -p "$(dirname "$dest")"
dest="$(cd "$(dirname "$dest")" && pwd)/$(basename "$dest")"

# 2. Clone, unless a primus checkout is already there.
if [ -f "$dest/$MAIN_MK_REL" ]; then
  echo "primus already present at $dest — skipping clone."
else
  echo ">> cloning signoz/primus into $dest"
  gh repo clone signoz/primus "$dest"
fi

# 3. Persist PRIMUS_HOME to the right shell profile (idempotent, non-clobbering).
case "$(basename "${SHELL:-sh}")" in
  zsh)  profile="${ZDOTDIR:-$HOME}/.zshrc" ;;
  bash) profile="$HOME/.bashrc" ;;
  *)    profile="$HOME/.profile" ;;
esac

line="export PRIMUS_HOME=\"$dest\""
if [ -f "$profile" ] && grep -qsF "export PRIMUS_HOME=" "$profile"; then
  echo "!! $profile already exports PRIMUS_HOME — left untouched. Intended value:"
  echo "     $line"
else
  printf '\n# SigNoz primus (terraform-provider-signoz tooling)\n%s\n' "$line" >> "$profile"
  echo ">> added to $profile:"
  echo "     $line"
fi

echo ">> activate now:  export PRIMUS_HOME=\"$dest\"   (or open a new shell / source $profile)"
echo ">> verify:        make -f \"$dest/$MAIN_MK_REL\" help"
