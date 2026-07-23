# Scalar Unions

Companion to [`patterns.md`](patterns.md) (the pattern catalog).
This is the decision behind the API value shapes that have **no faithful single
Terraform type**: scalar unions, scalar-or-array unions, and free-form objects.
The provider models all of them the same way — as an opaque JSON string
(`jsontypes.Normalized`).

## Problem

A Terraform attribute has exactly one type (`string` / `number` / `bool` / `list`
/ `object` / …). Several correct OpenAPI shapes map to no single type:

- **Scalar union** — a `oneOf` of scalars, e.g. `step: string | number` (the
  canonical `Querybuildertypesv5Step`, `$ref`'d at ~12 query sites). Occurs both
  as a named component and inline on a property (e.g. `FunctionArg.value:
  number | string`).
- **Scalar-or-array union** — `oneOf [string, array<string>]`, e.g. a list
  variable's `defaultValue`: a single default *or* a multi-select list.
- **Free-form object** — `{type: object}` with no `properties`, e.g.
  `SigNozDatasourceSpec`, which is always sent as `{}` (empty by design, not an
  oversight).

There is no lossless 1:1 mapping, so a representation must be **chosen**. Left
unhandled, each fails distinctly: a scalar/free-form value cascade-skips the
convertor step (`StringValue ↔ *…Union`, `ObjectValue ↔ map[string]interface{}`),
taking every parent with it; a scalar-or-array union is rejected outright by
tfplugingen-openapi (`unsupported multi-type`).

## What we decided, and why

**Model every such value as `jsontypes.Normalized`**, the framework-native
opaque-JSON string type — following AWS. One uniform mechanism for all shapes; the
API stays honest (`oneOf` / free-form object unchanged on the wire).

Why over the alternatives:

- **vs. plain-string + stringify numbers** — works for scalar
  unions whose string form is natural (durations), but doesn't generalize to
  object/array unions and needs per-field parse/format logic. `jsontypes` is one
  mechanism for every shape.
- **vs. forcing the wire type to `string`** (what a `property_types` override
  did) — **banned.** It changes the apitype on *both* sides, so a Read fails if
  the wire ever returns a number/object. `jsontypes` changes only the *schema*
  type; the apitype stays an honest union → no Read-failure risk. The provider's
  one such override (`FunctionArg.value → string`) was removed in favor of this.
- **vs. `types.Dynamic`** — rejected (above).

This was a deliberate, locked choice: apply the AWS-style jsontype uniformly across
the whole tree rather than tune representation per field.

**Accepted trade-off:** `jsontypes.Normalized` compares JSON structurally but does
**not** normalize number representation (`60` vs `60.0` still diffs), and string
members must be written with `jsonencode`. Uniformity across the tree was chosen
over per-field UX polish.

## How to do it

For the provider author this is mostly automatic — the shape is detected from the
spec and the `jsontypes.Normalized` typing + a raw-JSON convertor bridge are
generated. What you need to know per shape:

| Shape | Upstream (SigNoz) change? |
|-------|---------------------------|
| Named or inline scalar `oneOf` (`Step`, `FunctionArg.value`) | **None** — automatic. Any old `property_types: → string` override is removed. |
| Free-form `{type: object}` (`SigNozDatasourceSpec`) | **None** — automatic (an empty object is a legitimate shape). |
| Scalar-or-array `oneOf` (`defaultValue`) | **One:** expose it as a **named** `oneOf` component. An inline multi-type has no union to bridge and is rejected by tfplugingen-openapi; naming it is the faithful fix (forcing a single concrete type would drop one of the two forms). |

**Terraform UX** — the attribute value is JSON:

- a number → `step = "60"`
- a string member → `step = jsonencode("60s")` (bare `step = "60s"` is a JSON
  validation error)
- an empty free-form object → `spec = "{}"`
- a scalar-or-array → `default_value = jsonencode("prod")` **or**
  `default_value = jsonencode(["prod", "staging"])`

**Spotting it in a new resource:** if the pipeline reports
`StringValue ↔ *<Union>` or `ObjectValue ↔ map[string]interface{}` (and `<Union>`
is a scalar `oneOf`, or the target is a free-form object), it's this pattern — it
generates once the shape is seen; a scalar-or-array first needs the
named-component change above.

## External References

Surveyed AWS, Google, PagerDuty:

- **AWS** — models opaque/heterogeneous values as a **framework-native JSON-string custom type** (`jsontypes.Normalized`, `SmithyJSON[T]`) whose `StringSemanticEquals` compares JSON structurally (whitespace- and key-order-insensitive). This is the model we follow.
- **Google** — the SDKv2 version: `TypeString` + `validation.StringIsJSON` + `StateFunc: NormalizeJsonString`.
- **`types.Dynamic`** — *nobody* uses it for a union (poor plan ergonomics).