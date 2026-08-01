---
paths:
  - ".github/workflows/**"
---

# Comments

Do not write unnecessary comments where the code is self-explanatory. Document only non-obvious behavior, constraints, formats, and edge cases — never restate what the code already says.

Language rules build on this one: [`go-comments`](go-comments.md), [`py-comments`](py-comments.md).

## Workflows

A step's `name`, `uses`, and inputs already say what it does — don't narrate them. Comment only a constraint a future edit would silently break, and keep it to a line. Rationale that belongs in prose (why a version is pinned, why a job exists) goes in the README or the PR, not the YAML.
