---
paths:
  - "**/*.go"
---

# Go checks before opening a PR

For any change touching Go, run the primus Go checks locally **before opening a PR**. They mirror the CI jobs (`signoz/primus.workflows`), so green locally means green in CI. All four must pass.

```sh
make -f "$PRIMUS_HOME/src/make/main.mk" go-fmt    # format Go files
make -f "$PRIMUS_HOME/src/make/main.mk" go-lint   # lint (matches CI)
make -f "$PRIMUS_HOME/src/make/main.mk" go-deps   # verify go.mod / go.sum
make -f "$PRIMUS_HOME/src/make/main.mk" go-test   # run tests
```

Run them through primus, not a bare `golangci-lint` / `go test` — the targets pin the versions and flags CI uses. Needs `PRIMUS_HOME` set; see the "Tooling — primus" section in `.claude/CLAUDE.md` for one-time setup (check `PRIMUS_HOME` first — don't re-clone if it's already there).
