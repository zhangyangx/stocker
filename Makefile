BINARY_NAME=stocker
BUILD_DIR=build

.PHONY: all mac windows clean

all: mac windows

mac:
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd

windows:
	mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME).exe ./cmd

clean:
	rm -rf $(BUILD_DIR)