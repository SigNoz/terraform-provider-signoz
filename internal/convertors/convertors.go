// Package conv defines the contract types and interfaces that the
// generated conversion layer (`zz_generated_<x>.go`) and the
// hand-written hooks (`internal/service/<r>_hooks.go`) both adhere to.
//
// The contracts are *generic* — every concrete instantiation pins the
// framework-side type (`Tf`, `Model`) and the wire-side type (`Api`,
// `Dto`) at compile time. Generated code provides default implementations
// per customtype; hand-written hooks plug in alongside them at the points
// declared in `config.yml`.
//
// Two layers, two roles:
//
//	Layer 2 — per-customtype Expand / Flatten:
//	    Expander[Tf, Api]      Tf  → *Api
//	    Flattener[Tf, Api]     *Api → Tf
//
//	Layer 3 — resource-level hooks:
//	    PreExpander[Model]              Model    → Model
//	    PostExpander[Model, Dto]        *Dto, Model → *Dto
//	    PreFlattener[Dto]               *Dto     → *Dto
//	    PostFlattener[Model, Dto]       Model, *Dto → Model
//	    FieldExpander[Field, Wire]      Field    → Wire
//	    FieldFlattener[Field, Wire]     Wire     → Field
//
// Each interface has a `<X>Func` adapter so a free function can satisfy
// the interface without wrapping in a struct (Go's `http.HandlerFunc`
// pattern).
package conv

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// ---------------------------------------------------------------------------
// Layer 2 — per-customtype Expand / Flatten
// ---------------------------------------------------------------------------

// Expander converts a Terraform-framework-side typed value (Tf) — usually
// a `customtypes.<X>Value` — into the wire-format API DTO (`*Api`,
// usually `*apitypes.<X>`). Generated `Expand<X>` functions in
// `internal/conv/zz_generated_<x>.go` implicitly satisfy this contract;
// they're emitted as plain Go functions and adapted to the interface
// via `ExpanderFunc` only when a generic interface is needed (e.g. when
// composing or testing).
type Expander[Tf, Api any] interface {
	Expand(ctx context.Context, v Tf) (*Api, diag.Diagnostics)
}

// ExpanderFunc adapts a free function to satisfy Expander.
type ExpanderFunc[Tf, Api any] func(ctx context.Context, v Tf) (*Api, diag.Diagnostics)

// Expand satisfies Expander.
func (f ExpanderFunc[Tf, Api]) Expand(ctx context.Context, v Tf) (*Api, diag.Diagnostics) {
	return f(ctx, v)
}

// Flattener is the inverse of Expander: wire DTO → framework value.
type Flattener[Tf, Api any] interface {
	Flatten(ctx context.Context, in *Api) (Tf, diag.Diagnostics)
}

// FlattenerFunc adapts a free function to satisfy Flattener.
type FlattenerFunc[Tf, Api any] func(ctx context.Context, in *Api) (Tf, diag.Diagnostics)

// Flatten satisfies Flattener.
func (f FlattenerFunc[Tf, Api]) Flatten(ctx context.Context, in *Api) (Tf, diag.Diagnostics) {
	return f(ctx, in)
}

// ---------------------------------------------------------------------------
// Layer 3 — resource-level hooks
// ---------------------------------------------------------------------------

// PreExpander runs first in the Create/Update path. Takes the framework
// model (the user's HCL plan), returns a possibly-modified model that
// the generated expand will then walk. Use for: cross-field validation,
// fixed-value defaulting (`version = "v5"`), dropping fields the server
// auto-derives.
type PreExpander[Model any] interface {
	PreExpand(ctx context.Context, m Model) (Model, diag.Diagnostics)
}

// PreExpanderFunc adapts a free function to satisfy PreExpander.
type PreExpanderFunc[Model any] func(ctx context.Context, m Model) (Model, diag.Diagnostics)

// PreExpand satisfies PreExpander.
func (f PreExpanderFunc[Model]) PreExpand(ctx context.Context, m Model) (Model, diag.Diagnostics) {
	return f(ctx, m)
}

// PostExpander runs after the generated expand. Takes the API DTO and
// the source model (for cross-referencing), returns a possibly-modified
// DTO. Use for: dropping fields the server overrides, fixing wire-shape
// inconsistencies the generator can't infer.
type PostExpander[Model, Dto any] interface {
	PostExpand(ctx context.Context, in *Dto, src Model) (*Dto, diag.Diagnostics)
}

// PostExpanderFunc adapts a free function to satisfy PostExpander.
type PostExpanderFunc[Model, Dto any] func(ctx context.Context, in *Dto, src Model) (*Dto, diag.Diagnostics)

// PostExpand satisfies PostExpander.
func (f PostExpanderFunc[Model, Dto]) PostExpand(ctx context.Context, in *Dto, src Model) (*Dto, diag.Diagnostics) {
	return f(ctx, in, src)
}

// PreFlattener runs before the generated flatten. Takes the API DTO,
// returns a possibly-modified DTO. Use for: reconstructing decomposed
// envelopes (rule's `evaluation`), parsing opaque-data blobs
// (notification_channel's `data`), normalizing oneOf-of-primitives.
type PreFlattener[Dto any] interface {
	PreFlatten(ctx context.Context, g *Dto) (*Dto, diag.Diagnostics)
}

// PreFlattenerFunc adapts a free function to satisfy PreFlattener.
type PreFlattenerFunc[Dto any] func(ctx context.Context, g *Dto) (*Dto, diag.Diagnostics)

// PreFlatten satisfies PreFlattener.
func (f PreFlattenerFunc[Dto]) PreFlatten(ctx context.Context, g *Dto) (*Dto, diag.Diagnostics) {
	return f(ctx, g)
}

// PostFlattener runs after the generated flatten. Takes the framework
// model (just populated from the wire) and the source DTO (for
// cross-referencing), returns a possibly-modified model. Use for:
// canonicalizing server-normalized values (durations), cross-field
// defaulting that depends on the wire shape.
type PostFlattener[Model, Dto any] interface {
	PostFlatten(ctx context.Context, next Model, src *Dto) (Model, diag.Diagnostics)
}

// PostFlattenerFunc adapts a free function to satisfy PostFlattener.
type PostFlattenerFunc[Model, Dto any] func(ctx context.Context, next Model, src *Dto) (Model, diag.Diagnostics)

// PostFlatten satisfies PostFlattener.
func (f PostFlattenerFunc[Model, Dto]) PostFlatten(ctx context.Context, next Model, src *Dto) (Model, diag.Diagnostics) {
	return f(ctx, next, src)
}

// FieldExpander overrides the generated expand for ONE attribute. Use
// for: oneOf-of-primitives (`step` accepts string|number),
// custom-marshaled types (rule's `step` returning `json.RawMessage`),
// any per-field coercion the generator can't derive from the schema.
//
// `Field` is the framework-side type (e.g. `types.String`); `Wire` is
// whatever the API DTO field expects (often `*string`, sometimes
// `json.RawMessage`).
type FieldExpander[Field, Wire any] interface {
	Expand(ctx context.Context, v Field) (Wire, diag.Diagnostics)
}

// FieldExpanderFunc adapts a free function to satisfy FieldExpander.
type FieldExpanderFunc[Field, Wire any] func(ctx context.Context, v Field) (Wire, diag.Diagnostics)

// Expand satisfies FieldExpander.
func (f FieldExpanderFunc[Field, Wire]) Expand(ctx context.Context, v Field) (Wire, diag.Diagnostics) {
	return f(ctx, v)
}

// FieldFlattener is the inverse of FieldExpander.
type FieldFlattener[Field, Wire any] interface {
	Flatten(ctx context.Context, v Wire) (Field, diag.Diagnostics)
}

// FieldFlattenerFunc adapts a free function to satisfy FieldFlattener.
type FieldFlattenerFunc[Field, Wire any] func(ctx context.Context, v Wire) (Field, diag.Diagnostics)

// Flatten satisfies FieldFlattener.
func (f FieldFlattenerFunc[Field, Wire]) Flatten(ctx context.Context, v Wire) (Field, diag.Diagnostics) {
	return f(ctx, v)
}
