<!-- markdownlint-disable first-line-h1 no-inline-html -->
<a href="https://terraform.io">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset=".github/terraform_logo_dark.svg">
    <source media="(prefers-color-scheme: light)" srcset=".github/terraform_logo_light.svg">
    <img src=".github/terraform_logo_light.svg" alt="Terraform logo" title="Terraform" align="right" height="50">
  </picture>
</a>

# Terraform SigNoz Provider

The **SigNoz Provider** enables [Terraform](https://terraform.io) to manage [SigNoz](https://signoz.io) resources.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Documentation

To update the documentation edit the files in templates/ and then run make docs.
The files in docs/ are auto-generated and should not be updated manually.

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org)
installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the
provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.
