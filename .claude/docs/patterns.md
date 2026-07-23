# Patterns

Recognize a shape, apply the known fix. Two sides: **Terraform-side** patterns (how the provider models and converts) and **SigNoz-side** patterns (schema shapes that need an *upstream* fix so skaff can map them at all). skaff does no magic — everything it skips is one of the shapes below.

Each entry is *shape · pitfall · fix · example* (examples are real cases we hit). This is a reference, not a log; per-resource history lives in git. Background: [`architecture.md`](architecture.md) (the pipeline) and [`convertors.md`](convertors.md) (how model↔wire conversion works).

---

## Terraform-side patterns — how the provider models & converts

### Convertor shape

Every shipped resource is generated **field-for-field**: flatten returns `*schemas.<R>DataSourceModel` (the datasource is a superset of the resource), codegen narrows it with `<r>ResourceFromDS`, and `flex`/`services` print `skipped 0`. **There are no hand-written convertors** — `internal/convertors` and `internal/services` are 100% `zz_generated_*`. **Example:** route_policy, planned_maintenance, service_account, rule, dashboard — all declared in `skaff.yml` and generated end-to-end.

If a resource *can't* map field-for-field, that is **not** a convertor to hand-write — it's a **schema gap to fix upstream** (SigNoz-side patterns below). skaff won't emit a usable convertor, and hand-writing one would just paper over a wrong schema. See [`convertors.md`](convertors.md).

- **CRUD response-shape.** POST/PUT may return the full object, a partial `{id}`, or `204` — and the spec isn't always faithful to which. The generated shell defaults to a follow-up **GET** (POST→GET, PUT→GET) so state always lands via one Flatten path. **Example:** `rule`'s PUT returns `204`, so Update re-reads via GET.

### Values Terraform can't type 1:1 → `jsontypes.Normalized`

Scalar unions, scalar-or-array unions, and legitimately-opaque free-form objects. Full decision, rationale, and Terraform UX: [`scalar-unions.md`](scalar-unions.md).

- **Pitfall:** the convertor step reports `basetypes.StringValue ↔ *<Union>` (or `ObjectValue ↔ map[string]interface{}`) and every parent cascade-skips; a scalar-or-array is rejected outright (`unsupported multi-type`).
- **Fix:** model the attribute as `jsontypes.Normalized` — details in [`scalar-unions.md`](scalar-unions.md).

### Flattened-oneOf customtype ↔ union apitype

A discriminated (`kind`/`type`) `oneOf`: the customtype is a **flat object** (one optional field per variant); the wire type is an **oapi-codegen union**. skaff bridges them via the discriminator (`ValueByDiscriminator`).

- **Example:** dashboard `layouts`, panel plugins, variables — `dashboardtypes_layout` generates a flat `{ grid = {…} }` customtype bridged to the `{kind, spec}` union.

### Map-of-object (`additionalProperties: $ref`)

A `map[string]Obj` on the wire.

- **Pitfall:** skaff once tracked only `$ref` / `items.$ref`, so map *element* schemas never canonicalized (`<Prop>Type` / `<Prop>Value` dangled) → the map was reported "unsupported."
- **Fix:** skaff `MapValueRef` + generated `Expand/Flatten<X>Map` (handled now).
- **Example:** dashboard `spec.panels` (`map[string]Panel`) and `spec.datasources`.

---

## SigNoz-side patterns — schema shapes that need an upstream fix

Fix in the Go source that generates `docs/api/openapi.yml`, **not** in skaff, so the fix is permanent. The schema-shaping tools, roughly in order of reach: `required:"true"` / `nullable:"true"` struct tags · `Enum() []any` · `JSONSchema()` (the `jsonschema.Exposer` interface) · `JSONSchemaOneOf()` + `PrepareJSONSchema()` with the `x-signoz-discriminator` extension.

- **Untyped property** — `foo: {}` (no `type` / `$ref` / `oneOf`). *Pitfall:* tfplugingen-openapi errors *"no 'type' … attribute cannot be created"* → the whole enclosing schema is skipped. *Fix:* type the field (concrete `type` / `$ref` / `Enum()` / a `JSONSchema()` exposer). *Example:* `Querybuildertypesv5FunctionArg.value` was `value: {}` → added `PrepareJSONSchema` typing it `oneOf [number, string]` (signoz#11850).

- **Inline-nested object / inline `oneOf`** — a property (or array `items`) whose value is an object or `oneOf` with **no `$ref`**. *Pitfall:* `skaff types` *"has inline-nested object property (no $ref to canonicalize)"*, or tfplugingen *"schema composition is currently not supported"* → skipped. *Fix:* extract it into a named `#/components/schemas/X` and `$ref` it. *Example:* the builder-join aggregations were an inline `items.oneOf` → extracted into a named `Querybuildertypesv5JoinAggregation` component, which `detectOneOfRewrites` then flattens.

- **Untyped array items** — `items: {}`. *Pitfall:* the element has no type; the element attribute can't be created. *Fix:* type the `items` (`type:` or `$ref`). *Example:* `Querybuildertypesv5QueryData.items` (`[]any`) — a query-result type; type its items if a data source ever surfaces it.

- **Free-form object** — `{type: object}` with no `properties`. *Pitfall:* skaff models an empty-object customtype with no convertor → every field that `$ref`s it cascade-skips. *Fix:* give it real `properties` (a named struct) — or a discriminator if it's a sum. If it is **legitimately** opaque, model it as `jsontypes.Normalized` (Terraform-side, above). *Example:* `DashboardtypesSigNozDatasourceSpec` is always `{}` by design (schema owner confirmed) → opaque `jsontypes.Normalized` route; the user writes `spec = "{}"`.

- **By-`type`/`kind` sum object** — exposes `JSONSchemaOneOf()`, but the reflector also leaves the struct's own untyped base field (e.g. `spec: {}`) on the parent. *Pitfall:* the `oneOf` is mappable, but the leftover untyped base property re-trips the "untyped property" skip. *Fix:* add `PrepareJSONSchema()` (with `x-signoz-discriminator` when the discriminant maps 1:1 to variants) to strip the duplicate parent properties, leaving a clean `oneOf`-of-`$ref` for `detectOneOfRewrites`. *Example:* `Querybuildertypesv5QueryEnvelope.spec` — `PrepareJSONSchema` dropped the reflected base `spec: {}`, leaving the clean query-variant union (signoz#11850). Related gotcha: each variant's `type` had to be `required:"true"` so it renders **non-pointer** — oapi-codegen's `From<Variant>` can't assign a const string to a `*T`.

- **Untyped map** — `additionalProperties: {}`. *Pitfall:* skaff auto-coerces it to `{type: string}` — compiles, but silently lossy if the values aren't strings. *Fix:* type `additionalProperties` properly (usually a `$ref` → the map-of-object pattern above). *Example:* dashboard `spec.panels` first appeared as `additionalProperties: {}` (coerced to string) before it was typed to a panel `$ref`.

- **Nullable number → `float32` precision drift** — a `*float64` Go field emits bare `{type: number}` with **no `format`** (swaggest only adds `format` for non-pointer floats). *Pitfall:* oapi-codegen maps bare `number` → `float32`, so `target = 0.8` round-trips to `0.800000011920929` and apply fails: *"produced an unexpected new value: … was cty…(\"0.8\"), but now …(0.800000011920929)"*. *Fix:* add a `` `format:"double"` `` struct tag → `float64` everywhere. *Example:* `signoz_rule` `condition.thresholds.basic.spec[].{target,recoveryTarget}` and `condition.target` (signoz#12061).

- **Association via path params** — identity/CRUD expressed through path params (`/service_accounts/{id}/roles`, delete `/roles/{rid}`), a list-returning read, and no update endpoint. *Pitfall:* skaff builds the model from request-body ⊕ response (never path params), so the model degenerates to a bare `{id}`, `flex` skips (list read, no single-object GET), and `services` has no update — not Pattern A. *Fix:* give the association its own first-class endpoints (`POST /<association>` → `{id}`, `GET`/`DELETE /<association>/{id}`) so identity and CRUD live in the body. Details: [`association.md`](association.md). *Example:* `service_account_role` (`POST /service_accounts/{id}/roles` + list read + `DELETE …/roles/{rid}`, no update).

- **`204 No Content` that declares a body** — the spec generator emits a `content` block for a 204 whenever a response content-type is set, even when there's no body (`Response == nil`). *Pitfall:* oapi-codegen then generates an unconditional `json.Unmarshal` for the 204; the server correctly returns 204 with an empty body → `unexpected end of JSON input` (e.g. on delete). The server is right — only the spec is wrong. *Fix:* in the SigNoz spec generator (`pkg/http/handler/handler.go` `ServeOpenAPI`), omit the content type when `Response == nil` — one `else` branch fixes the whole class (18 bodyless responses; the per-route `ResponseContentType: ""` was the targeted first pass). Landing it is a **3-repo regen**: fix the *generator source* (not the generated `openapi.yml`) → regen the spec (`go run ./cmd/community generate openapi`) → regen the frontend client (`pnpm generate:api`, orval) → regen the provider (skaff). *Example:* a bodyless `DELETE` returning 204 → `unexpected end of JSON input`.
