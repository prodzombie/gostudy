.DEFAULT_GOAL := help

GO ?= go
ARGS ?=

.PHONY: help build run demo test fmt vet check clean

help:
	@printf '%s\n' \
	  'make build              Build bin/gostudy' \
	  'make run ARGS="review"  Run the CLI with arguments' \
	  'make demo               Review a temporary copy of the examples' \
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

demo:
	@set -eu; \
	demo_dir=$$(mktemp -d); \
	trap 'rm -rf "$$demo_dir"' EXIT; \
	trap 'exit 130' INT; \
	trap 'exit 143' TERM; \
	cp -R examples/cards "$$demo_dir/"; \
	$(GO) run . -d "$$demo_dir" review -t gostudy

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
