.PHONY: all
all: client server

client: $(wildcard cmd/client/*.go) $(wildcard internal/**/*.go)
	go build ./cmd/client

server: $(wildcard cmd/server/*.go) $(wildcard internal/**/*.go)
	go build ./cmd/server

.PHONY: check
check:
	@if [ -n "$$(gofmt -l .)" ]; \
	then \
		echo 'Formatting issues detected!'; \
		echo 'Run `make format` to fix.'; \
		exit 1; \
	else \
		echo 'No formatting issues!'; \
	fi
	go test -cover ./...

.PHONY: clean
clean:
	rm -f client server

.PHONY: format
format:
	go fmt ./...
