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

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.4, or [OpenTofu](https://opentofu.org/docs/intro/install/) >= 1.6
- [Go](https://golang.org/doc/install) >= 1.25 (only to build the provider from source)

### Why Terraform 1.4 and not 1.0

The provider is built on the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework), and models much of the SigNoz API as *nested attributes* — objects inside objects, several levels deep, whose inner fields are both optional and computed: you may set them, and the server supplies whatever you leave out.

Terraform 1.3 and older do not round-trip that shape. After a successful apply, the values the server filled in for attributes the configuration left unset are not reconciled with the plan, and the next `plan` proposes them as changes all over again. The effect is perpetual drift — the apply succeeds, and every subsequent plan wants to make the identical change. No change to the configuration avoids it.

Resources whose attributes are all flat scalars are unaffected, so an older CLI looks fine right up until a deeply nested resource is used. Terraform 1.4 is the first release that handles it. OpenTofu inherits the fix, its first release having been forked from Terraform 1.5.

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
