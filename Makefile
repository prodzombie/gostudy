.DEFAULT_GOAL := help

GO ?= go
ARGS ?=

.PHONY: help build run test fmt vet check clean

help:
	@printf '%s\n' \
	  'make build              Build bin/gostudy' \
	  'make run ARGS="review"  Run the CLI with arguments' \
	  'make test               Run all tests' \
	  'make fmt                Format Go source files' \
	  'make vet                Check for common Go mistakes' \
	  'make check              Run tests and vet' \
	  'make clean              Remove the built CLI'

build:
	mkdir -p bin
	$(GO) build -o bin/gostudy .

run:
	$(GO) run . $(ARGS)

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

check: test vet

clean:
	$(GO) clean
	rm -f bin/gostudy
