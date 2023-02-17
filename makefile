.PHONY: test lint

test:
	go test -v -cover ./...

lint:
	golangci-lint run ./...

lintfixer:
	golangci-lint run --fix ./...
