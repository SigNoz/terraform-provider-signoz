BINARY        := terraform-provider-signoz
GOPATH        := $(shell go env GOPATH)
GOBIN         := $(GOPATH)/bin
INSTALL_PATH  := $(GOBIN)/$(BINARY)

.PHONY: build install dev-setup test testacc lint docs fmt vet clean

## build: compile the provider binary to ./bin/
build:
	go build -o bin/$(BINARY) .

## install: install the provider binary to $GOPATH/bin (required for dev_overrides)
install:
	go install .

## dev-setup: write ~/.terraformrc with dev_overrides pointing to $GOPATH/bin
dev-setup:
	@echo 'Writing ~/.terraformrc with dev_overrides...'
	@printf 'provider_installation {\n  dev_overrides {\n    "registry.terraform.io/SigNoz/signoz" = "%s"\n  }\n  direct {}\n}\n' "$(GOBIN)" > ~/.terraformrc
	@echo 'Done. Run "make install" to build and place the binary.'

## fmt: run gofmt on all Go source files
fmt:
	gofmt -s -w .

## vet: run go vet
vet:
	go vet ./...

## lint: run go vet + staticcheck (install: go install honnef.co/go/tools/cmd/staticcheck@latest)
lint: vet
	@which staticcheck > /dev/null 2>&1 || (echo "staticcheck not found: go install honnef.co/go/tools/cmd/staticcheck@latest" && exit 1)
	staticcheck ./...

## test: run unit tests
test:
	go test ./... -v -count=1

## docs: regenerate provider docs from templates
docs:
	go generate ./...

## clean: remove local build artifacts
clean:
	rm -rf bin/

## help: list available targets
help:
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
