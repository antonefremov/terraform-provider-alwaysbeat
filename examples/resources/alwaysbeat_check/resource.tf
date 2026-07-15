# A cron-scheduled check whose ping URL feeds straight into the job that runs it.
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

# A simple interval check with run-too-long alerting.
resource "alwaysbeat_check" "hourly_sync" {
  name = "hourly-sync"

  schedule = {
    kind     = "interval"
    interval = "1h"
    tz       = "UTC"
  }

  grace        = "5m"
  max_run      = "10m"
  max_run_mode = "hung"
  channels     = ["webhook:https://example.com/hooks/alwaysbeat"]
}

output "backup_ping_url" {
  value = alwaysbeat_check.nightly_backup.ping_url
}
