package main

// Code generation for the provider's registry documentation. Run with
// `go generate ./...` from the repository root (Terraform must be on PATH so
// the example configurations can be formatted).

// Format the example Terraform configurations under examples/.
//go:generate terraform fmt -recursive ./examples/

// Regenerate docs/ from the provider schema, templates/, and examples/.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate -provider-name signoz -rendered-provider-name SigNoz
