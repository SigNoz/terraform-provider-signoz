# Association

Some SigNoz endpoints model a **link between two existing objects** rather than a standalone object with its own lifecycle — `service_account_role` (role ↔ service account) and `user_role` (role ↔ user). These are association / membership resources (the Terraform analogue of `aws_iam_role_policy_attachment`).

skaff can't generate a usable resource for the *current* API shape. The fix is **not** to hand-write it or change skaff — it's to **model the association as a first-class resource in the SigNoz API** (new endpoints) so skaff maps it field-for-field like any other Pattern-A resource.

## Why skaff can't map it — identity lives in path params

The association's identity and operations are expressed through **path parameters**, not a self-contained object. For `service_account_role`:

- create: `POST /api/v1/service_accounts/{id}/roles`, body `{ "id": <roleId> }`
- read:   `GET  /api/v1/service_accounts/{id}/roles` → a **list** of roles
- delete: `DELETE /api/v1/service_accounts/{id}/roles/{rid}` (two path params)
- no update endpoint

skaff (via `tfplugingen-openapi`) builds the Terraform model from the **create request body ⊕ the read response** — never from path parameters. So:

- **The composite identity is lost.** The parent `{id}` (service account) is only a path param, so it can't be made `Required`/`ForceNew`, and it collides with the body's `id` (the role) — the model degenerates to a single `{ id }`. The delete's `{rid}` never reaches the schema at all.
- **The read has no single object to flatten.** It returns a *list*, not a `GET …/{id}` of one object, so there is no response data type → the convertor step (`flex`) is skipped and no Expand/Flatten is generated.

None of this is Pattern A (one object, one self-`id`, single-object GET, full CRUD), so the identity a Terraform resource needs simply isn't in the parts of the spec skaff reads.

## The fix — give the association its own endpoints in SigNoz

Model the association as a **standalone resource** in the SigNoz API, so its identity and CRUD live in the request/response body (what skaff maps) instead of in path parameters. Add endpoints of the form:

- `POST /api/v1/<association>` — body carries both foreign keys (e.g. `{ service_account_id, role_id }`) and returns the created object `{ id, … }`.
- `GET /api/v1/<association>/{id}` — returns the single object by its own id.
- `DELETE /api/v1/<association>/{id}`.

Now the association is an ordinary object with its own `id` and single-object CRUD, so skaff generates it **field-for-field**, like any other resource — no hand-written convertor, no skaff change. Declare it in `skaff.yml` and run the pipeline.
