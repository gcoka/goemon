
.PHONY: example
example: build-example
	make -C example start

build-example:
	go build -o example/goemon main.go

.PHONY: test
test:
	go test -race ./...

lint:
	go vet ./...
	go list ./... | xargs golint -set_exit_status
