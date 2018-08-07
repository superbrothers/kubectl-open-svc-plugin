PLUGIN_NAME := open-svc
OUT_DIR := ./_out
DIST_DIR := ./_dist
$(shell mkdir -p _dist)

.PHONY: build
build:
		go build -o $(PLUGIN_NAME) .

.PHONY: cross
build-cross: $(OUT_DIR)/linux-amd64/$(PLUGIN_NAME) $(OUT_DIR)/darwin-amd64/$(PLUGIN_NAME)

.PHONY: dist
dist: $(DIST_DIR)/$(PLUGIN_NAME)-linux-amd64.zip $(DIST_DIR)/$(PLUGIN_NAME)-darwin-amd64.zip

.PHONY: checksum
checksum:
		for f in _dist/*.zip; do \
			shasum -a 256 "$${f}"  | awk '{print $$1}' > "$${f}.sha256" ; \
		done

.PHONY: clean
clean:
		rm -rf $(OUT_DIR) $(DIST_DIR) $(PLUGIN_NAME)

$(OUT_DIR)/%-amd64/$(PLUGIN_NAME):
		GOOS=$* GOARCH=amd64 go build -o $@ .

$(DIST_DIR)/$(PLUGIN_NAME)-%-amd64.zip: $(OUT_DIR)/%-amd64/$(PLUGIN_NAME)
		( \
			cd $(OUT_DIR)/$*-amd64/ && \
			cp ../../version.txt . && \
			cp ../../LICENSE . && \
			cp ../../README.md . && \
			cp ../../plugin.yaml . && \
			zip -r ../../$@ * \
		)
