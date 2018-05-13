
.PHONY: example
example:
	$(MAKE) build-example
	make -C example start

build-example:
	go build -o example/goemon main.go
