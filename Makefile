.PHONY: build
build:
	go build -o build/tatakae ./cmd/tatakae

.PHONY: test
test:
	go test ./...
