GO ?= go
DIST_DIR := dist

.PHONY: build
build:
	$(GO) build -o $(DIST_DIR)/kubectl-open_svc cmd/kubectl-open_svc.go

TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin
$(shell mkdir -p $(TOOLS_BIN_DIR))

GORELEASER_BIN := bin/goreleaser
GORELEASER := $(TOOLS_DIR)/$(GORELEASER_BIN)
GORELEASER_VERSION ?= v1.8.3
GOLANGCI_LINT_BIN := bin/golangci-lint
GOLANGCI_LINT := $(TOOLS_DIR)/$(GOLANGCI_LINT_BIN)
VALIDATE_KREW_MAIFEST_BIN := bin/validate-krew-manifest
VALIDATE_KREW_MAIFEST := $(TOOLS_DIR)/$(VALIDATE_KREW_MAIFEST_BIN)

$(GORELEASER):
	curl -SL "https://github.com/goreleaser/goreleaser/releases/download/$(GORELEASER_VERSION)/goreleaser_$(shell uname -s)_$(shell uname -m).tar.gz" | tar xz -C $(TOOLS_BIN_DIR) goreleaser

$(GOLANGCI_LINT): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && $(GO) build -o $(GOLANGCI_LINT_BIN) github.com/golangci/golangci-lint/cmd/golangci-lint

$(VALIDATE_KREW_MAIFEST): $(TOOLS_DIR)/go.mod
	cd $(TOOLS_DIR) && $(GO) build -o $(VALIDATE_KREW_MAIFEST_BIN) sigs.k8s.io/krew/cmd/validate-krew-manifest

.PHONY: build-cross
build-cross: $(GORELEASER)
	$(GORELEASER) build --snapshot --rm-dist

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: lint
lint: vet fmt $(GOLANGCI_LINT)
	$(GOLANGCI_LINT) run

.PHONY: test
test:
	$(GO) test -v ./...

.PHONY: validate-krew-manifest
validate-krew-manifest: $(VALIDATE_KREW_MAIFEST)
	$(VALIDATE_KREW_MAIFEST) -manifest dist/open-svc.yaml -skip-install

.PHONY: dist
dist: $(GORELEASER)
	$(GORELEASER) release --rm-dist --skip-publish --snapshot

.PHONY: release
release: $(GORELEASER)
	$(GORELEASER) release --rm-dist --skip-publish

.PHONY: clean
clean:
	rm -rf $(DIST_DIR) $(TOOLS_BIN_DIR)
