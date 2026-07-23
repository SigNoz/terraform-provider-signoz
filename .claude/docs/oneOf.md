# `oneOf`

How a `oneOf` (a value that is exactly one of several variants) maps into a Terraform schema. For the specific *scalar/opaque* `oneOf` cases, see [`scalar-unions.md`](scalar-unions.md); this doc is the general shape decision.

## Problem

Terraform's type system has no `oneOf`/`anyOf`. An attribute is exactly one concrete type. So a polymorphic API object (e.g. a query that is *one of* builder-trace / builder-log / builder-metric / formula / promql / clickhouse-sql / trace-operator) has no direct representation — you must choose an encoding and enforce "exactly one variant" yourself.

## What we decided, and why

**Nested attributes, not blocks.** Reasons: HashiCorp recommends attributes for new framework schemas; blocks can't live under attributes (our `oneOf`s are list elements and nested data, not top-level sub-blocks); and our codegen path already emits attributes. Blocks are reserved for a future hand-authored DSL resource where HCL ergonomics clearly win.

Concretely, three variant shapes:

- **Discriminated object `oneOf`** (by `kind`/`type`) → **one optional `SingleNestedAttribute` per variant** on a parent object, exactly-one enforced. In a list, a `ListNestedAttribute` whose element carries the sibling variants (order preserved). skaff's `detectOneOfRewrites` flattens a `oneOf`-of-`$ref` into exactly this shape and bridges it to the wire union ([`patterns.md`](patterns.md) → "flattened-oneOf ↔ union").
- **Scalar / scalar-or-array `oneOf`** (`string | number`, `string | [string]`) → `jsontypes.Normalized` ([`scalar-unions.md`](scalar-unions.md)).
- **`types.Dynamic` / raw `any`** → avoided: punts validation to runtime and loses schema help.

**Example.** A rule's `condition.composite_query.queries[]` is an ordered list of `Querybuildertypesv5QueryEnvelope` (7 variants). It maps to a `ListNestedAttribute` where each element sets exactly one of `builder_query` / `builder_formula` / `promql` / `clickhouse_sql` / `builder_trace_operator`, ordering preserved — authored as `queries = [{ promql = {…} }, { builder_query = {…} }]`.

## How it lands here

Discriminated `oneOf`-of-`$ref` components are flattened automatically by skaff (the expand/flatten bridge maps the chosen sibling back to the wire `{kind|type, spec}` union). The upstream requirement is a clean `oneOf`-of-`$ref` with the duplicate untyped base property stripped — see [`patterns.md`](patterns.md) "by-`type`/`kind` sum object".

## External References

- **HashiCorp guidance** — for *new* framework schemas, use **nested attributes**, not blocks; block support exists mainly to migrate legacy SDK providers. And blocks can't sit underneath attributes, so a deep `oneOf` inside an attribute-shaped object would force the whole parent chain to be blocks.
- **codegen issue #94** — the maintainer's worked example models each variant as an optional sibling `SingleNestedAttribute` plus a `ExactlyOneOf` validator; blocks are acknowledged as *possible*, not prescribed.
- **AWS** — uses **blocks** for complex mutually-exclusive object alternatives (WAFv2 statement trees, Bedrock knowledge-base configs) and **attributes** for simpler/primitive choices. Real precedent for blocks, but on hand-written resources.
- **PingOne** — historically block-based (`oidc_options`/`saml_options`); current source has moved to `SingleNestedAttribute` + `objectvalidator.ExactlyOneOf`.
