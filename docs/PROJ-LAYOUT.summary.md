# Project Layout Summary

```
terraform-provider-signoz/
├── signoz/                         # Provider source
│   ├── internal/attr/              #   Schema attributes
│   ├── internal/client/            #   API client
│   ├── internal/model/             #   Data models
│   ├── internal/provider/datasource/ # Data sources
│   ├── internal/provider/resource/ #   Resources
│   ├── internal/utils/             #   Utilities
│   └── provider.go                 #   Provider factory
├── docs/                           # Registry docs
├── templates/                      # Doc templates
├── examples/                       # HCL examples
├── tools/                          # Go tool deps
├── .github/workflows/              # CI/CD
├── main.go                         # Entry point
├── go.mod / go.sum                 # Go module
├── .goreleaser.yml                 # Release config
└── terraform-registry-manifest.json
```
