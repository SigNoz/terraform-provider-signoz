<!-- markdownlint-disable first-line-h1 no-inline-html -->
<a href="https://terraform.io">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset=".github/terraform_logo_dark.svg">
    <source media="(prefers-color-scheme: light)" srcset=".github/terraform_logo_light.svg">
    <img src=".github/terraform_logo_light.svg" alt="Terraform logo" title="Terraform" align="right" height="50">
  </picture>
</a>

# Terraform SigNoz Provider

The **SigNoz Provider** lets [Terraform](https://terraform.io) manage [SigNoz](https://signoz.io) observability resources as code, on both SigNoz Cloud and self-hosted deployments.

📖 **[Provider documentation on the Terraform Registry](https://registry.terraform.io/providers/signoz/signoz/latest/docs)**

## Usage

```terraform
terraform {
  required_providers {
    signoz = {
      source = "signoz/signoz"
    }
  }
}

provider "signoz" {
  # SigNoz Cloud region URL, or the UI URL of a self-hosted deployment. Also reads
  # SIGNOZ_ENDPOINT; defaults to http://localhost:8080.
  endpoint = "http://localhost:8080"

  # API access token from a service account. Prefer the SIGNOZ_ACCESS_TOKEN
  # environment variable to keep the secret out of configuration and state.
  access_token = var.signoz_access_token
}
```

Create an access token from a [SigNoz service account](https://signoz.io/docs/manage/administrator-guide/iam/service-accounts/).
See the [registry documentation](https://registry.terraform.io/providers/signoz/signoz/latest/docs) for every resource, data source, and example.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.25 (only to build the provider from source)

## Building the provider

1. Clone the repository.
2. Enter the repository directory.
3. Build the provider with the Go `install` command:

```shell
go install
```

## Adding dependencies

This provider uses [Go modules](https://go.dev/wiki/Modules). To add a new
dependency `github.com/author/dependency`:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Documentation

Documentation is generated with
[terraform-plugin-docs](https://github.com/hashicorp/terraform-plugin-docs) from the
provider schema, the prose templates in `templates/`, and the examples in `examples/`.
Edit those sources, then regenerate:

```shell
go generate ./...
```

The files in `docs/` are auto-generated — do not edit them by hand. CI re-runs
generation and fails if `docs/` is out of date.

## Developing the provider

You'll need [Go](https://go.dev/doc/install) >= 1.25 (see [Requirements](#requirements)).
To compile the provider, run `go install`; the binary lands in `$GOPATH/bin`. After
changing the schema, examples, or templates, run `go generate ./...` to refresh `docs/`.
