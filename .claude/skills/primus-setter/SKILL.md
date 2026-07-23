---
name: primus-setter
description: >-
  Set up SigNoz primus for this repo — clone it to a location the user picks (default `.primus` at the repo root) and export `PRIMUS_HOME`. Use when primus is not set up: `PRIMUS_HOME` is unset, `make -f "$PRIMUS_HOME/…" help` fails, or the user says "set up primus" / "install primus". Checks first and is a no-op if primus is already set up — never re-clones over an existing install.
---

# Set up primus

primus (`signoz/primus`, private) provides the repo's `skaff` codegen and the Go checks (`go-fmt`/`go-lint`/`go-deps`/`go-test`). This skill installs it and sets `PRIMUS_HOME`. Requires `gh` with SigNoz org access.

## Steps

1. **Check first — is it already set up?** If `PRIMUS_HOME` is exported and valid, primus is installed; **stop, do not set up again.**

```sh
[ -n "$PRIMUS_HOME" ] && make -f "$PRIMUS_HOME/src/make/main.mk" help >/dev/null 2>&1 \
   && echo "already set up — stop" || echo "not set up — continue"
```

2. **Ask the user where to set up primus.** Take their answer as the clone location. If they don't give one (empty), default to `.primus` at the repo root (gitignored).

3. **Run the setup script** with that location (empty → default):

```sh
bash .claude/skills/primus-setter/scripts/setup.sh "<location-or-empty>"
```

It clones `signoz/primus` there and appends `export PRIMUS_HOME="<location>"` to the user's shell profile (`.zshrc` / `.bashrc` / `.profile`), idempotently — it won't re-clone or overwrite an existing `PRIMUS_HOME` export.

4. **Tell the user to activate it** — `export PRIMUS_HOME="<location>"` in the current shell, or open a new shell. Verify: `make -f "$PRIMUS_HOME/src/make/main.mk" help`.
