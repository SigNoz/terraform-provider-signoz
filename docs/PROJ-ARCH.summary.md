# Architecture Summary

Terraform provider for SigNoz (alerts + dashboards) built on `terraform-plugin-framework` v1.14 (protocol v6).

**Components**: Provider → API Client (heimdall, retry+backoff) → SigNoz REST API (SIGNOZ-API-KEY auth).

**Resources**: `signoz_alert` (CRUD via `/api/v2/rules`), `signoz_dashboard` (CRUD via `/api/v1/dashboards`). Both also available as data sources (read-only).

**Config**: `access_token` (required), `endpoint`, `http_timeout`, `http_max_retry` — settable via HCL or env vars (`SIGNOZ_ACCESS_TOKEN`, `SIGNOZ_ENDPOINT`, etc.).

**Stack**: Go 1.25, heimdall HTTP client, GoReleaser, GitHub Actions CI, published to `registry.terraform.io/signoz/signoz`.
