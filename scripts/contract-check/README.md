# contract-check

Proves the provider's wire shapes still match the SigNoz server's request and
response types. `main.go` marshals the bodies the provider emits and decodes
them into the server's strict request types (`DisallowUnknownFields` plus the
server's own `Validate`), then marshals the server's response types and decodes
them into the provider's client shapes the same way. Covered: `log_pipeline`
(`pipelinetypes.PostablePipelines` + `IsValid`), `data_retention`
(`retentiontypes` v2 POST body and GET response), `plugin_dashboard`
(`dashboardtypes` Postable/Updatable `Validate`, Gettable round-trip including
spec-level and per-panel links), and the gateway ingestion key/limit
(`gatewaytypes` bodies and the `{"status","data"}` envelopes).

The provider must never depend on the server module, so the check runs inside a
throwaway module in `mktemp -d` whose go.mod `replace`s both sides — the
provider and the server checkout — and imports the provider's `internal/model`
through its module path (allowed: the throwaway module path is a subpath of the
provider's).

## Run

```sh
make contract-check                       # SIGOZ_DIR defaults to ../../signoz
SIGOZ_DIR=~/src/signoz make contract-check
```

`main.go` carries a `//go:build ignore` tag so the parent module never builds
it; `run.sh` strips the tag when copying it into the throwaway module. The temp
directory is removed on exit. A non-zero exit means at least one contract check
failed — fix the provider (or update the check alongside a deliberate contract
change on both sides) before shipping.
