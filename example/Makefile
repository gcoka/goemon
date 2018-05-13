BIN_NAME=example_bin
BUILD_CMD_LOCAL=go build -o $(BIN_NAME) test.go

.PHONY: start
start:
	./goemon -x "pkill $(BIN_NAME) & ($(BUILD_CMD_LOCAL)) && ./$(EXE_NAME)"
