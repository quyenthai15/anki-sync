.PHONY: help setup build dist dist-mac dist-win skill-zip check profiles dry-run import clean

BINARY_NAME=anki-sync
SKILL_NAME=anki-sync
SKILL_DIR=.agents/skills/$(SKILL_NAME)
BIN_DIR=$(SKILL_DIR)/bin
DIST_DIR=./dist
SRC_DIR=./cmd/sync
DRAFT?=.anki-draft.yaml

help:
	@echo "Setup:"
	@echo "  make setup              Create .profiles.json if missing and print setup steps"
	@echo ""
	@echo "Build:"
	@echo "  make build              Build ./anki-sync for this machine"
	@echo "  make dist               Build macOS and Windows binaries into $(BIN_DIR)/ and the skill zip into dist/"
	@echo "  make skill-zip          Zip the skill (with bin/) for Claude Desktop into dist/"
	@echo ""
	@echo "Usage (uses DRAFT=<file>, default $(DRAFT)):"
	@echo "  make check              Check AnkiConnect is running"
	@echo "  make profiles           List language profiles"
	@echo "  make dry-run            Validate the draft without writing to Anki"
	@echo "  make import             Import the draft into Anki"
	@echo ""
	@echo "Other:"
	@echo "  make clean              Remove built binaries, bin/ and dist/"

setup:
	@if [ ! -f .profiles.json ]; then cp .profiles.example.json .profiles.json && echo "Created .profiles.json"; else echo ".profiles.json already exists"; fi
	@echo ""
	@echo "Setup steps:"
	@echo "  1. Install Anki desktop and the AnkiConnect add-on (code 2055492159)."
	@echo "  2. In Anki: File > Import > data/fr_sample.apkg  (once only; re-import duplicates notes)."
	@echo "  3. Open Anki, then run: anki-sync check (or $(BIN_DIR)/<platform>/anki-sync)."
	@echo "  4. Run: anki-sync profiles"
	@echo "  5. Install the skill: copy $(SKILL_DIR) to ~/.claude/skills/ (or upload dist/$(SKILL_NAME).zip to Claude Desktop)."

build:
	go build -o $(BINARY_NAME) $(SRC_DIR)

dist: dist-mac dist-win skill-zip
	@echo "Skill zip: $(DIST_DIR)/$(SKILL_NAME).zip"

dist-mac:
	@mkdir -p $(BIN_DIR)/darwin-arm64
	GOOS=darwin GOARCH=arm64 go build -o $(BIN_DIR)/darwin-arm64/$(BINARY_NAME) $(SRC_DIR)

dist-win:
	@mkdir -p $(BIN_DIR)/windows-amd64
	GOOS=windows GOARCH=amd64 go build -o $(BIN_DIR)/windows-amd64/$(BINARY_NAME).exe $(SRC_DIR)

# Claude Desktop expects a zip whose single top-level folder is the skill name.
skill-zip: dist-mac dist-win
	@mkdir -p $(DIST_DIR)
	@rm -f $(DIST_DIR)/$(SKILL_NAME).zip
	cd .agents/skills && zip -qr ../../$(DIST_DIR)/$(SKILL_NAME).zip $(SKILL_NAME) -x '*.DS_Store'

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
	rm -rf $(BIN_DIR) $(DIST_DIR)
	@echo "Cleaned"
