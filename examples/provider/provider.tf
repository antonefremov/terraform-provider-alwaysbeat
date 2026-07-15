terraform {
  required_providers {
    alwaysbeat = {
      source = "antonefremov/alwaysbeat"
    }
  }
}

provider "alwaysbeat" {
  # endpoint defaults to the production API; override for staging/local:
  # endpoint = "https://staging.example"
  #
  # Prefer the ALWAYSBEAT_API_KEY environment variable over setting api_key here,
  # to keep the key out of config and state.
  # api_key = var.alwaysbeat_api_key
}
