# Changelog

## 0.1.1

BUG FIXES:

* data-source/`alwaysbeat_check`: fix a `Value Conversion Error` ("Received null
  value ... Path: schedule") when reading a check. The data source now reads only
  `id` from config and builds the remaining (computed) attributes from the API,
  instead of decoding null nested attributes into non-nullable structs. The
  `alwaysbeat_check` resource was unaffected.

## 0.1.0

FEATURES:

* **New Resource:** `alwaysbeat_check` — manage a cron/heartbeat check as code,
  with full CRUD, `terraform import`, and a computed `ping_url`.
* **New Data Source:** `alwaysbeat_check` — look up an existing check by `id`.

Initial release: manage [Alwaysbeat](https://alwaysbeat.com) monitoring checks
with Terraform. Durations use Go duration strings (`"5m"`, `"1h30m"`); the API
is the source of truth for validation.
