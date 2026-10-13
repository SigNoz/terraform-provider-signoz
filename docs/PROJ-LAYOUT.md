# Project Layout

```
terraform-provider-signoz/
├── signoz/                         # Provider source code
│   ├── internal/
│   │   ├── attr/                   #   Terraform schema attribute definitions
│   │   ├── client/                 #   SigNoz REST API client
│   │   ├── model/                  #   Go structs for alert & dashboard models
│   │   ├── provider/
│   │   │   ├── datasource/         #     Data source implementations
│   │   │   └── resource/           #       Resource implementations
│   │   └── utils/                  #   Shared utility functions
│   └── provider.go                 #   Provider factory (New function)
├── docs/                           # Terraform Registry documentation
│   ├── data-sources/               #   Data source reference (alert, dashboard)
│   ├── resources/                  #   Resource reference (alert, dashboard)
│   └── index.md                    #   Provider index page
├── templates/                      # Doc generation templates (tfplugindocs)
│   ├── data-sources/
│   └── resources/
├── examples/                       # HCL examples for docs and testing
│   ├── data-sources/               #   Data source examples
│   ├── resources/                  #   Resource examples
│   ├── existing-alert/             #   Import existing alert example
│   ├── existing-dashboard/         #   Import existing dashboard example
│   ├── install-verification/       #   Smoke test config
│   └── provider/                   #   Provider configuration example
├── tools/                          # Go tool dependencies (tfplugindocs)
│   └── tools.go
├── .github/                        # CI/CD and GitHub config
│   ├── workflows/
│   │   ├── release.yml             #     GoReleaser-based release pipeline
│   │   └── test.yml                #     Test and lint workflow
│   ├── dependabot.yml              #     Dependency update config
│   └── CODEOWNERS
├── main.go                         # Provider entry point
├── go.mod                          # Go module definition
├── go.sum                          # Dependency checksums
├── .goreleaser.yml                 # GoReleaser config for binary builds
├── .golangci.yml                   # Linter configuration
├── terraform-registry-manifest.json # Terraform Registry protocol version
├── .gitignore
├── LICENSE
└── README.md                       # Start here
```

## Key Files

| File | Purpose |
|------|---------|
| `main.go` | Entry point — calls `providerserver.Serve` |
| `signoz/provider.go` | Provider factory — registers resources and data sources |
| `signoz/internal/client/client.go` | HTTP client for SigNoz API |
| `.goreleaser.yml` | Cross-platform build matrix for releases |
| `terraform-registry-manifest.json` | Declares Terraform protocol version (6) |
