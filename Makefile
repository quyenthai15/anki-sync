.PHONY: help build dist dist-mac dist-win check profiles dry-run import clean

BINARY_NAME=anki-sync
DIST_DIR=./dist
SRC_DIR=./cmd/sync
DRAFT?=.anki-draft.yaml

help:
	@echo "Build:"
	@echo "  make build              Build ./anki-sync for this machine"
	@echo "  make dist               Build macOS (Apple Silicon) and Windows (64-bit) into dist/"
	@echo ""
	@echo "Usage (uses DRAFT=<file>, default $(DRAFT)):"
	@echo "  make check              Check AnkiConnect is running"
	@echo "  make profiles           List language profiles"
	@echo "  make dry-run            Validate the draft without writing to Anki"
	@echo "  make import             Import the draft into Anki"
	@echo ""
	@echo "Other:"
	@echo "  make clean              Remove built binaries and dist/"

build:
	go build -o $(BINARY_NAME) $(SRC_DIR)

dist: dist-mac dist-win
	@echo "Binaries are in $(DIST_DIR)/"

dist-mac:
	@mkdir -p $(DIST_DIR)/mac-arm
	GOOS=darwin GOARCH=arm64 go build -o $(DIST_DIR)/mac-arm/$(BINARY_NAME) $(SRC_DIR)

dist-win:
	@mkdir -p $(DIST_DIR)/win64
	GOOS=windows GOARCH=amd64 go build -o $(DIST_DIR)/win64/$(BINARY_NAME).exe $(SRC_DIR)

check: build
	./$(BINARY_NAME) check

profiles: build
	./$(BINARY_NAME) profiles

dry-run: build
	./$(BINARY_NAME) import --dry-run $(DRAFT)

import: build
	./$(BINARY_NAME) import $(DRAFT)

clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe sync
	rm -rf $(DIST_DIR)
	@echo "Cleaned"
