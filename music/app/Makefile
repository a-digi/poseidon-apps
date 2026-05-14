MAKEFILE_DIR := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
PLUGIN_DIR   := $(abspath $(MAKEFILE_DIR)/..)
DIST_DIR     := $(CURDIR)/dist
PLUGIN_VERSION := $(shell awk -F'"' '/"version"/ {print $$4; exit}' $(PLUGIN_DIR)/manifest.json | tr '.' '-')

.DEFAULT_GOAL := all

.PHONY: all
all: backend frontend

.PHONY: backend
backend:
	cd $(MAKEFILE_DIR)/backend && go build -o $(PLUGIN_DIR)/backend/music .
	cd $(MAKEFILE_DIR)/backend && go vet ./...

.PHONY: frontend
frontend:
	cd $(MAKEFILE_DIR)/frontend && npm run build

.PHONY: zip
zip: zip-all

.PHONY: clean
clean:
	rm -f $(PLUGIN_DIR)/backend/music
	rm -rf $(PLUGIN_DIR)/ui/*
	rm -rf $(DIST_DIR)
	rm -rf $(MAKEFILE_DIR)/.build

# ── Cross-platform builds ─────────────────────────────────────────────────────

PLUGIN_ID := music

.PHONY: backend-darwin-arm64
backend-darwin-arm64:
	mkdir -p $(MAKEFILE_DIR)/.build/darwin-arm64
	cd $(MAKEFILE_DIR)/backend && \
		CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 \
		go build -o $(MAKEFILE_DIR)/.build/darwin-arm64/$(PLUGIN_ID) .

.PHONY: backend-darwin-amd64
backend-darwin-amd64:
	mkdir -p $(MAKEFILE_DIR)/.build/darwin-amd64
	cd $(MAKEFILE_DIR)/backend && \
		CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
		go build -o $(MAKEFILE_DIR)/.build/darwin-amd64/$(PLUGIN_ID) .

.PHONY: backend-linux-amd64
backend-linux-amd64:
	mkdir -p $(MAKEFILE_DIR)/.build/linux-amd64
	cd $(MAKEFILE_DIR)/backend && \
		CGO_ENABLED=1 CC='zig cc -target x86_64-linux-musl' GOOS=linux GOARCH=amd64 \
		go build -o $(MAKEFILE_DIR)/.build/linux-amd64/$(PLUGIN_ID) .

.PHONY: backend-windows-amd64
backend-windows-amd64:
	mkdir -p $(MAKEFILE_DIR)/.build/windows-amd64
	cd $(MAKEFILE_DIR)/backend && \
		CGO_ENABLED=1 CC='zig cc -target x86_64-windows-gnu' GOOS=windows GOARCH=amd64 \
		go build -o $(MAKEFILE_DIR)/.build/windows-amd64/$(PLUGIN_ID).exe .

.PHONY: zip-darwin-arm64
zip-darwin-arm64: backend-darwin-arm64 frontend
	mkdir -p $(DIST_DIR)
	STAGE=$$(mktemp -d) && \
	mkdir -p $$STAGE/backend $$STAGE/ui && \
	cp $(PLUGIN_DIR)/manifest.json $$STAGE/ && \
	cp $(MAKEFILE_DIR)/.build/darwin-arm64/$(PLUGIN_ID) $$STAGE/backend/$(PLUGIN_ID) && \
	cp -R $(PLUGIN_DIR)/ui/. $$STAGE/ui/ && \
	(cd $$STAGE && zip -r $(DIST_DIR)/$(PLUGIN_ID)_$(PLUGIN_VERSION)_darwin-arm64.zip .) && \
	rm -rf $$STAGE

.PHONY: zip-darwin-amd64
zip-darwin-amd64: backend-darwin-amd64 frontend
	mkdir -p $(DIST_DIR)
	STAGE=$$(mktemp -d) && \
	mkdir -p $$STAGE/backend $$STAGE/ui && \
	cp $(PLUGIN_DIR)/manifest.json $$STAGE/ && \
	cp $(MAKEFILE_DIR)/.build/darwin-amd64/$(PLUGIN_ID) $$STAGE/backend/$(PLUGIN_ID) && \
	cp -R $(PLUGIN_DIR)/ui/. $$STAGE/ui/ && \
	(cd $$STAGE && zip -r $(DIST_DIR)/$(PLUGIN_ID)_$(PLUGIN_VERSION)_darwin-amd64.zip .) && \
	rm -rf $$STAGE

.PHONY: zip-linux-amd64
zip-linux-amd64: backend-linux-amd64 frontend
	mkdir -p $(DIST_DIR)
	STAGE=$$(mktemp -d) && \
	mkdir -p $$STAGE/backend $$STAGE/ui && \
	cp $(PLUGIN_DIR)/manifest.json $$STAGE/ && \
	cp $(MAKEFILE_DIR)/.build/linux-amd64/$(PLUGIN_ID) $$STAGE/backend/$(PLUGIN_ID) && \
	cp -R $(PLUGIN_DIR)/ui/. $$STAGE/ui/ && \
	(cd $$STAGE && zip -r $(DIST_DIR)/$(PLUGIN_ID)_$(PLUGIN_VERSION)_linux-amd64.zip .) && \
	rm -rf $$STAGE

.PHONY: zip-windows-amd64
zip-windows-amd64: backend-windows-amd64 frontend
	mkdir -p $(DIST_DIR)
	STAGE=$$(mktemp -d) && \
	mkdir -p $$STAGE/backend $$STAGE/ui && \
	cp $(PLUGIN_DIR)/manifest.json $$STAGE/ && \
	cp $(MAKEFILE_DIR)/.build/windows-amd64/$(PLUGIN_ID).exe $$STAGE/backend/$(PLUGIN_ID).exe && \
	cp -R $(PLUGIN_DIR)/ui/. $$STAGE/ui/ && \
	(cd $$STAGE && zip -r $(DIST_DIR)/$(PLUGIN_ID)_$(PLUGIN_VERSION)_windows-amd64.zip .) && \
	rm -rf $$STAGE

.PHONY: zip-all
zip-all: zip-darwin-arm64 zip-darwin-amd64 zip-linux-amd64 zip-windows-amd64
