---
paths:
  - "**/*.go"
  - ".github/workflows/**"
---

# Comments

Do not write unnecessary comments where the code is self-explanatory.

- **Godoc**: Skip comments that merely restate the identifier. Document only non-obvious behavior, constraints, formats, and edge cases.
- **Generated code**: If the comment is emitted by an external codegen tool, leave it as-is — do not add or trim comments in generated files.
- **Workflows**: Same bar. A step's `name`, `uses`, and inputs already say what it does — don't narrate them. Comment only a constraint a future edit would silently break, and keep it to a line. Rationale that belongs in prose (why a version is pinned, why a job exists) goes in the README or the PR, not the YAML.
