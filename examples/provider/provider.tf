terraform {
  required_providers {
    stillbeat = {
      source = "antonefremov/stillbeat"
    }
  }
}

provider "stillbeat" {
  # endpoint defaults to the production API; override for staging/local:
  # endpoint = "https://staging.example"
  #
  # Prefer the STILLBEAT_API_KEY environment variable over setting api_key here,
  # to keep the key out of config and state.
  # api_key = var.stillbeat_api_key
}
