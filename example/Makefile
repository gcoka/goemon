BIN_NAME=example_bin
BUILD_CMD_LOCAL=go build -o $(BIN_NAME) server.go

.PHONY: start
start:
	./goemon --config nodemon.json --print -v "$(BUILD_CMD_LOCAL) && ./$(BIN_NAME)"
