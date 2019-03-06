PLUGIN_NAME := kubectl-open_svc
OUT_DIR := ./_out
OUTS := $(OUT_DIR)/linux-amd64/$(PLUGIN_NAME) $(OUT_DIR)/darwin-amd64/$(PLUGIN_NAME)
DIST_DIR := ./_dist
DISTS := $(DIST_DIR)/$(PLUGIN_NAME)-linux-amd64.zip $(DIST_DIR)/$(PLUGIN_NAME)-darwin-amd64.zip
CHECKSUMS := $(DISTS:.zip=.zip.sha256)

$(shell mkdir -p _dist)

.PHONY: build
build:
		go build -o $(PLUGIN_NAME) cmd/$(PLUGIN_NAME).go

build-cross: $(OUTS)

.PHONY: vet
vet:
		go tool vet -printfuncs Infof,Warningf,Errorf,Fatalf,Exitf pkg

.PHONY: fmt
fmt:
		go fmt ./pkg/... ./cmd/...

dist: $(DIST_DIR)/open-svc.yaml

test-dist: dist
		./hack/test-dist.sh

$(OUT_DIR)/%-amd64/$(PLUGIN_NAME):
		GOOS=$* GOARCH=amd64 go build -o $@ cmd/$(PLUGIN_NAME).go

$(DIST_DIR)/$(PLUGIN_NAME)-%-amd64.zip.sha256: $(DIST_DIR)/$(PLUGIN_NAME)-%-amd64.zip
		shasum -a 256 "$^"  | awk '{print $$1}' > "$@"

$(DIST_DIR)/$(PLUGIN_NAME)-%-amd64.zip: $(OUT_DIR)/%-amd64/$(PLUGIN_NAME)
		( \
			cd $(OUT_DIR)/$*-amd64/ && \
			cp ../../LICENSE . && \
			cp ../../README.md . && \
			zip -r ../../$@ * \
		)

$(DIST_DIR)/open-svc.yaml: $(DISTS) $(CHECKSUMS)
		./hack/generate-plugin-yaml.sh >"$@"

.PHONY: clean
clean:
		rm -rf $(OUT_DIR) $(DIST_DIR) $(PLUGIN_NAME)
