---
paths:
  - "**/*.py"
---

# Python comments

The bar is the [`comments`](comments.md) rule: nothing where the code is self-explanatory.

- **No file-level docstring.** The filename says what the module is for — `tool_bin.py` gets the tool binary. A module docstring restating that is noise, and a paragraph of design prose at the top of a file goes stale where nobody is looking. A constraint belongs next to the code it constrains, not in a preamble.
- **Docstrings**: only when they say something the name and signature don't — drop them otherwise. Keep them short. A contract that genuinely needs a few lines (interacting flags, retry semantics, an edge case) is fine; a narrative is not.
- **No song and dance.** Comment the constraint or the edge case. Not the narrative, not the rationale, not what the next line does.
