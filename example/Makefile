BIN_NAME=example_bin
BUILD_CMD_LOCAL=go build -o $(BIN_NAME) test.go

.PHONY: start
start:
	./goemon --config goemon.yml "$(BUILD_CMD_LOCAL) && ./$(BIN_NAME)"
