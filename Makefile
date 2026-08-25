BINARY_NAME=sota
BUILD_DIR=bin

build:
	@echo "Building Sota..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

release-all:
	@echo "Building binaries for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "Done! Check the '$(BUILD_DIR)' directory."

clean:
	rm -rf $(BUILD_DIR)
	@echo "Cleaned up!"