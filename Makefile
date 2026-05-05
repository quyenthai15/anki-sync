.PHONY: build dist clean

BINARY_NAME=sync
DIST_DIR=../anki-sync/dist/go
SRC_DIR=./cmd/sync

# Default build for current platform
build:
	go build -o $(BINARY_NAME) $(SRC_DIR)

# Colors and Emojis
BLUE    = \033[34m
GREEN   = \033[32m
CYAN    = \033[36m
BOLD    = \033[1m
RESET   = \033[0m

# Distribution builds for Win/Mac
dist:
	@rm -rf $(DIST_DIR)
	@echo "Removed old files"
	@mkdir -p $(DIST_DIR)
	@echo "$(BLUE)📦 Starting distribution builds...$(RESET)"

# @echo "$(BLUE)  🔨 Building for macOS (Intel)...$(RESET)"
# @mkdir -p $(DIST_DIR)/mac-intel
# @GOOS=darwin GOARCH=amd64 go build -o $(DIST_DIR)/mac-intel/$(BINARY_NAME) $(SRC_DIR)

	@echo "$(BLUE)  🔨 Building for macOS (Apple Silicon)...$(RESET)"
	@mkdir -p $(DIST_DIR)/mac-arm
	@GOOS=darwin GOARCH=arm64 go build -o $(DIST_DIR)/mac-arm/$(BINARY_NAME) $(SRC_DIR)

	@echo "$(BLUE)  🔨 Building for Windows (64-bit)...$(RESET)"
	@mkdir -p $(DIST_DIR)/win64
	@GOOS=windows GOARCH=amd64 go build -o $(DIST_DIR)/win64/$(BINARY_NAME).exe $(SRC_DIR)

	@echo "\n$(GREEN)$(BOLD)✅ All builds completed successfully!$(RESET)"
	@echo "$(GREEN)📍 Binaries are located in: $(DIST_DIR)$(RESET)"

clean:
	rm -f $(BINARY_NAME)
	rm -rf $(DIST_DIR)
	@echo "Cleanup completed"
