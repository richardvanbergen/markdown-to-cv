# Makefile for m2cv - Markdown to CV

BINARY_NAME := m2cv
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-X github.com/richq/m2cv/cmd.version=$(VERSION) -X github.com/richq/m2cv/cmd.commit=$(COMMIT) -X github.com/richq/m2cv/cmd.date=$(DATE)"

.PHONY: all build install test clean

all: clean build

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

install:
	go install $(LDFLAGS) .

test:
	go test ./... -v

clean:
	rm -f $(BINARY_NAME)
