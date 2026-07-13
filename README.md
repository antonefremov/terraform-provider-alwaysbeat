# Terraform Provider for Stillbeat

Manage [Stillbeat](https://stillbeat.app) cron / heartbeat checks as
code. Define your monitored jobs in Terraform; wire the generated `ping_url`
straight into the cron/CI job that pings it.

> Status: **v0.1.0 (early)** — the `stillbeat_check` resource with full CRUD + import.
> Data sources and additional resources are planned.

## Usage

```hcl
terraform {
  required_providers {
    stillbeat = {
      source  = "antonefremov/stillbeat"
      version = "~> 0.1"
    }
  }
}

provider "stillbeat" {
  # api_key via STILLBEAT_API_KEY (recommended); endpoint defaults to production.
}

resource "stillbeat_check" "nightly_backup" {
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
  value = stillbeat_check.nightly_backup.ping_url
}
```

Then, in the job:

```sh
run-my-backup && curl -fsS "$(terraform output -raw ping_url)"
```

## Authentication

Create an API key in the dashboard under **API keys**, then either:

- set `STILLBEAT_API_KEY=dmf_...` in the environment (preferred — keeps it out of
  config and state), or
- pass `api_key` in the `provider "stillbeat"` block.

## Provider configuration

| Argument   | Env           | Default            | Description                                   |
|------------|---------------|--------------------|-----------------------------------------------|
| `api_key`  | `STILLBEAT_API_KEY` | —                  | Stillbeat API key (`dmf_...`). Required.            |
| `endpoint` | `STILLBEAT_ENDPOINT`| production API URL | API base URL; override for staging/local.     |

## `stillbeat_check`

Durations are Go duration strings (`"30s"`, `"5m"`, `"1h30m"`). Import with:

```sh
terraform import stillbeat_check.nightly_backup <check_id>
```

See [`examples/`](./examples) and the generated [`docs/`](./docs).

## Development

```sh
make build    # go install
make test     # unit tests (no network)
make testacc  # acceptance tests — set TF_ACC=1, STILLBEAT_API_KEY, STILLBEAT_ENDPOINT (STAGING!)
make docs     # regenerate docs/ (requires tfplugindocs)
```

Acceptance tests create and destroy real checks — always point `STILLBEAT_ENDPOINT`
at **staging**, never production.

## License

[MPL-2.0](./LICENSE).
