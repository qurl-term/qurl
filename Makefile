GO  ?= go
BIN := bin/qurl

.PHONY: build test race vet lint tidy check run clean

build:
    CGO_ENABLED=0 $(GO) build -trimpath -ldflags '-s -w' -o $(BIN) ./cmd/qurl

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

lint:
	golangci-lint run

tidy:
	$(GO) mod tidy

check: vet lint race

run: build
	./$(BIN) $(ARGS)

clean:
	rm -rf bin dist
