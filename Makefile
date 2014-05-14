SOURCE := $(wildcard *.go)

export GOPATH=$(shell pwd)

all:

test:
	@go test -v .

benchmark:
	@go test -v -bench=.

format: ${SOURCE}
	@gofmt -w ${SOURCE}

.PHONY: test format
