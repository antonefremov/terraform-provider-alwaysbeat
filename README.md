# Terraform Provider for Alwaysbeat

Manage [Alwaysbeat](https://alwaysbeat.com) cron / heartbeat checks as
code. Define your monitored jobs in Terraform; wire the generated `ping_url`
straight into the cron/CI job that pings it.

> Status: **v0.1.0 (early)** — the `alwaysbeat_check` resource with full CRUD + import.
> Data sources and additional resources are planned.

## Usage

```hcl
terraform {
  required_providers {
    alwaysbeat = {
      source  = "antonefremov/alwaysbeat"
      version = "~> 0.1"
    }
  }
}

provider "alwaysbeat" {
  # api_key via ALWAYSBEAT_API_KEY (recommended); endpoint defaults to production.
}

resource "alwaysbeat_check" "nightly_backup" {
  name = "nightly-backup"

  schedule = {
    kind      = "cron"
    cron_expr = "0 3 * * *"
    tz        = "Europe/Berlin"
  }

  grace    = "30m"
  channels = ["email:ops@example.com"]
}

output "ping_url" {
  value = alwaysbeat_check.nightly_backup.ping_url
}
```

Then, in the job:

```sh
run-my-backup && curl -fsS "$(terraform output -raw ping_url)"
```

## Authentication

Create an API key in the dashboard under **API keys**, then either:

- set `ALWAYSBEAT_API_KEY=dmf_...` in the environment (preferred — keeps it out of
  config and state), or
- pass `api_key` in the `provider "alwaysbeat"` block.

## Provider configuration

| Argument   | Env           | Default            | Description                                   |
|------------|---------------|--------------------|-----------------------------------------------|
| `api_key`  | `ALWAYSBEAT_API_KEY` | —                  | Alwaysbeat API key (`dmf_...`). Required.            |
| `endpoint` | `ALWAYSBEAT_ENDPOINT`| production API URL | API base URL; override for staging/local.     |

## `alwaysbeat_check`

Durations are Go duration strings (`"30s"`, `"5m"`, `"1h30m"`). Import with:

```sh
terraform import alwaysbeat_check.nightly_backup <check_id>
```

See [`examples/`](./examples) and the generated [`docs/`](./docs).

## Development

```sh
make build    # go install
make test     # unit tests (no network)
make testacc  # acceptance tests — set TF_ACC=1, ALWAYSBEAT_API_KEY, ALWAYSBEAT_ENDPOINT (STAGING!)
make docs     # regenerate docs/ (requires tfplugindocs)
```

Acceptance tests create and destroy real checks — always point `ALWAYSBEAT_ENDPOINT`
at **staging**, never production.

## License

[MPL-2.0](./LICENSE).
