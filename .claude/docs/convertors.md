# Convertors

How the provider converts between the Terraform model and the API wire types. The convertors are also **generated** — there is no hand-written conversion layer.

## Problem

A Terraform resource's schema model and the API's request/response types are always distinct Go types. So each resource needs code to convert both ways — handling null/pointer, snake<->camel, and nested/polymorphic shapes.

## What we decided

**Per-path code generation, via skaff — and nothing hand-written.** skaff emits one `Expand<X>` / `Flatten<X>` per OpenAPI component (customtype), called by the generated service CRUD shell. Conversion scales with **type count, not resource count**. Reflection is overkill at this scale; pure hand-writing doesn't scale.

**There is no hand-written convertor layer.**

> If a resource can't be generated **field-for-field**, that's a **schema violation** — fix the SigNoz schema so skaff *can* generate it. Don't hand-write a convertor.

Everything skaff can't map is a schema shape to fix upstream — see the SigNoz-side patterns in [`patterns.md`](patterns.md).

## How it works — two generation steps

The convertor layer (`internal/convertors/`, package `conv`) is emitted by **two** skaff commands, both writing there:

- **`skaff flex`** — one **per-resource entry point** for each field-for-field resource: `Expand<RequestBodyType>` (resource model → request apitype) and `Flatten<ResponseDataType>` (response apitype → `<R>DataSourceModel`), into `zz_generated_<resource>_flex.go`. These are what the generated `internal/services` CRUD shell calls. A resource that isn't field-for-field is **skipped** — which is the schema gap the rule above refers to (we require `skipped 0`).

- **`skaff convertors`** — one **per-customtype** `Expand<X>` / `Flatten<X>` for every OpenAPI component present in *both* the customtypes dir and the apitypes dir (`zz_generated_<x>.go`), recursing into nested customtypes. It dispatches on the `(framework-type, wire-type)` pair to pick the right primitive helper (`StringPointer`, `TimePointerFromString`, …).

Create/Update end with a follow-up GET so state always lands via one Flatten path (the CRUD response-shape note in [`patterns.md`](patterns.md)).

## Reachability pruning — why `convertors` runs last

`skaff convertors --callers-dir <services>` runs a **function-level reachability filter**, so it emits only the per-customtype helpers actually used:

1. Build a call graph over the `conv` package.
2. **Roots** = bare-name `conv.<X>` calls in the external `--callers-dir` (the generated `services` layer), **plus** the `flex` entry points (`zz_generated_*_flex.go`) — `flex` bridges `services` into the per-customtype helpers, so its files are additional edge sources / roots.
3. BFS from the roots; a schema is **reachable** iff at least one of its planned `Expand` / `Flatten` / `ValueFromObject` helpers is reached.
4. Emit `zz_generated_<x>.go` only for reachable schemas; prune stale files whose schema dropped out of the set.

Without `--callers-dir`, every planned schema is emitted. Because the roots include the `flex` and `services` output, `convertors` must run **last** in the pipeline (after both) — otherwise the reachable set is empty.

## External References

Surveyed AWS, PagerDuty, Google — three production patterns:

| Pattern | Who | Upfront | Per-resource |
|---------|-----|---------|--------------|
| Hand-written `buildX`/`flattenX` per resource | PagerDuty | none | high |
| **Per-path code-generated** `expand`/`flatten` | Google (magic-modules) | a generator + IDL | low |
| Reflective auto-marshal | AWS (AutoFlex, ~21k LOC) | large library | ~zero |

- **PagerDuty** hand-write one `expand`/`flatten` pair per nested type in the resource file.
- **Google** emits one `expand`/`flatten` per nested path from a YAML IDL.
- **AWS AutoFlex** reflectively walks the model; a new resource is ≈ a one-line `Expand`/`Flatten` call, but it costs ~21k LOC of reflection — only pays off at ~50+ resources.
- Nobody uses `types.Dynamic` as the primary interface.
