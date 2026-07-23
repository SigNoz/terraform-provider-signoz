# Opening pull requests

Open PRs with the `gh` CLI against `SigNoz/terraform-provider-signoz`.

- **Run the checks first.** For a Go change, run the primus `go-fmt` / `go-lint` / `go-deps` / `go-test` gates (see the `go-checks` rule); for docs/examples, the `docs-writer` gates. Don't open a PR on red.
- **Follow the template** (`.github/pull_request_template.md`): put each change under the right heading (Features / Fixes / Chores / Refactors / Tests) and drop the headings that don't apply. Don't add sections the template doesn't have.
- **Keep the description minimal and human-readable.** A few plain bullets saying what changed and why, written for a reviewer skimming it — not a wall of text, not a restatement of the diff, not generated boilerplate.
- **No AI attribution** in the title or body.
