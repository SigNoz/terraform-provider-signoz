---
paths:
  - "**/*.py"
---

# pytest conventions

For the Python integration suite under `tests/`.

- **No module-level `_helper()` functions.** Inline the logic into the test body — a reader should see what a test does without chasing private helpers, which scatter a test's meaning across the file. Prefer inline expressions, comprehensions, lists of `ids`, and pytest fixtures over `_`-prefixed functions. Extract a helper only when it's **genuinely non-trivial and reused across many tests**, or when explicitly asked.
- **Fixture-factory over indirect parametrization.** A fixture that returns a callable (e.g. `workspace(tf_file)`) is clearer than `@pytest.mark.parametrize(..., indirect=True)` + `request.param` — the value is an explicit argument, not resolved by magic.
- **Skip at collection, not inside the test body.** Use `pytest.param(..., marks=pytest.mark.skip(reason="…"))` so a skipped case shows as SKIPPED-with-reason **and** short-circuits before its fixtures run (no environment spin-up for a test that won't execute).
- **Test config comes from explicit `--flags`, not the environment.** Wire configuration as pytest options passed by the workflow; do **not** add`os.environ` fallbacks inside the tests.

## Gotchas

- **`testpaths` scopes a bare run; explicit paths override it.** Set `testpaths = ["integration/tests"]` so `pytest` collects only the suites, while `pytest integration/bootstrap/setup.py::test_setup` still runs the bootstrap entrypoints on demand. (`--ignore` in `addopts` would *also* block the explicit path — don't use it for this.)
- **Absolute-path option value → wrong `rootdir`.** Passing an absolute path as an *option value* (`--foundry-binary-path /tmp/foundryctl`) with no path arg makes pytest treat that path as a collection target -> `rootdir` resolves there -> `conftest.py` never loads -> the option reads as "unrecognized." Always pass a repo-rooted path arg (e.g. `integration/tests`) so `rootdir` resolves.
