default: build

# Build and install the provider into the local Go bin.
build:
	go install

# Unit tests (no network / no Terraform).
test:
	go test ./... -timeout=120s

# Acceptance tests — hit a REAL DMF API. Requires:
#   TF_ACC=1  DMF_API_KEY=dmf_...  [DMF_ENDPOINT=https://staging...]
# Point these at STAGING, never prod: acceptance tests create and destroy checks.
testacc:
	TF_ACC=1 go test ./... -v -timeout=120m

# Regenerate docs/ from schema + examples (requires tfplugindocs).
docs:
	go generate ./...

fmt:
	gofmt -w .

.PHONY: default build test testacc docs fmt
