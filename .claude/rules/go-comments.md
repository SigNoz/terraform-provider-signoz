---
paths:
  - "**/*.go"
---

# Comments

Do not write unnecessary comments where the code is self-explanatory.

- **Godoc**: Skip comments that merely restate the identifier. Document only non-obvious behavior, constraints, formats, and edge cases.
- **Generated code**: If the comment is emitted by an external codegen tool, leave it as-is — do not add or trim comments in generated files.
