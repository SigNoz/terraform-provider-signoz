# Architecture

The provider is **generated, not hand-written**: skaff reads the SigNoz OpenAPI spec + `skaff.yml` and emits the `internal/**` tree. This is the pipeline overview; each deep topic has its own doc (linked at the bottom).

## Why codegen and skaff?

Hand-writing didn't scale - the early hand-written pass was ~2,200 LOC across five resources, most of it mechanical schema-walking and expand/flatten boilerplate and **riddled with bugs**. With an OpenAPI spec and a generator, that mechanical layer is generated and scales with **type count, not resource count** and almost no bugs.

## The layers

skaff emits one `internal/` subdir per layer; each consumes the one below.

```
customtypes → schemas → apitypes → apiclients → convertors → services → provider
```

| Layer | Holds |
|-------|-------|
| `internal/customtypes/` | TF custom types for nested objects |
| `internal/schemas/` | resource/data-source schema + model structs |
| `internal/apitypes/` | oapi-codegen wire DTOs |
| `internal/apiclients/` | typed HTTP client methods |
| `internal/convertors/` | Expand (model→wire) / Flatten (wire→model) |
| `internal/services/` | the resource / data-source CRUD impls |
| `internal/provider/` | framework entry; registers each resource |

## Deep dives

- **Model ↔ wire conversion** (generated; includes the provider survey) → [`convertors.md`](convertors.md)
- **`oneOf` → Terraform** (nested attributes) → [`oneOf.md`](oneOf.md)
- **Audit/timestamp fields** (drop them) → [`audit-fields.md`](audit-fields.md)
- **Pattern catalog + schema fixes** → [`patterns.md`](patterns.md)
- **Opaque values** (`jsontypes.Normalized`) → [`scalar-unions.md`](scalar-unions.md)
- **Association resources** → [`association.md`](association.md)
