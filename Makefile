GO ?= go
DIST_DIR := dist

TOOLS_BIN_DIR := $(CURDIR)/hack/tools/bin
$(shell mkdir -p $(TOOLS_BIN_DIR))

GORELEASER := $(TOOLS_BIN_DIR)/goreleaser
GORELEASER_VERSION ?= v1.17.2
GOLANGCI_LINT := $(TOOLS_BIN_DIR)/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.52.2
VALIDATE_KREW_MAIFEST := $(TOOLS_BIN_DIR)/validate-krew-manifest
VALIDATE_KREW_MAIFEST_VERSION ?= v0.4.3
GORELEASER_FILTER_VERSION ?= v0.3.0
GORELEASER_FILTER := $(TOOLS_BIN_DIR)/goreleaser-filter

$(GORELEASER):
	GOBIN=$(TOOLS_BIN_DIR) $(GO) install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION)

$(GOLANGCI_LINT):
	GOBIN=$(TOOLS_BIN_DIR) $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

$(VALIDATE_KREW_MAIFEST):
	GOBIN=$(TOOLS_BIN_DIR) $(GO) install sigs.k8s.io/krew/cmd/validate-krew-manifest@$(VALIDATE_KREW_MAIFEST_VERSION)

$(GORELEASER_FILTER):
	GOBIN=$(TOOLS_BIN_DIR) go install github.com/t0yv0/goreleaser-filter@$(GORELEASER_FILTER_VERSION)

.PHONY: build
build: $(GORELEASER)
	$(GORELEASER) build --snapshot --clean --single-target --output $(DIST_DIR)/kubectl-open_svc

.PHONY: build-cross
build-cross: $(GORELEASER)
	$(GORELEASER) build --snapshot --clean

.PHONY: lint
lint: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run

.PHONY: lint-fix
lint-fix: $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run --fix

.PHONY: test
test:
	$(GO) test -v ./...

.PHONY: validate-krew-manifest
validate-krew-manifest: $(VALIDATE_KREW_MAIFEST)
	$(VALIDATE_KREW_MAIFEST) -manifest dist/open-svc.yaml -skip-install

.PHONY: dist
dist: $(GORELEASER_FILTER) $(GORELEASER)
	cat .goreleaser.yaml | $(GORELEASER_FILTER) -goos $(shell go env GOOS) -goarch $(shell go env GOARCH) | $(GORELEASER) release -f- --clean --skip-publish --snapshot

.PHONY: dist-all
dist-all: $(GORELEASER)
	$(GORELEASER) release --clean --skip-publish --snapshot

.PHONY: release
release: $(GORELEASER)
	$(GORELEASER) release --clean --skip-publish

.PHONY: clean
clean:
	rm -rf $(DIST_DIR) $(TOOLS_BIN_DIR)
