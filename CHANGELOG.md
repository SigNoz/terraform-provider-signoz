# Changelog

## 2026-10-13

- Rebased the fork onto upstream `main` @ `4a205dd` (v0.1.5 lineage): fork resources ported as hand-written services in `internal/services`, transport unified under `apiclients.WrappedClient` (new `DoRequest`), old `signoz/` tree dropped.
- Resource decisions: ours-only `alert`, `plugin_dashboard`, `downtime_schedule`, `data_retention`, `ingestion_key`, `log_pipeline` ported unchanged; collided `dashboard`, `notification_channel`, `route_policy`, `saved_view` keep the fork's legacy-API schemas (upstream's generated new-API versions stay in the tree unregistered, `NewLegacy*` constructors); upstream-only `auth_domain`, `planned_maintenance`, `role`, `rule`, `service_account`, `service_account_role`, `user`, `user_role` come in with the base.
- Adopt upstream v2 for the collided names: `signoz_dashboard`, `signoz_notification_channel`, `signoz_route_policy`, `signoz_saved_view` now register upstream's generated resources/data sources; the fork's legacy implementations (`internal/services/legacy_*`, `internal/legacyclient/{notification_channel,route_policy,saved_view,downtime_schedule}.go`, their models/attrs/tests/templates/examples) are deleted and `signoz_downtime_schedule` (snake_case keys the server never read) is dropped in favour of `signoz_planned_maintenance` — see `docs/guides/migrate-fork-legacy-resources.md`.
