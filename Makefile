SOURCE := $(wildcard *.go)

export GOPATH=$(shell pwd)

all:

test:
	@go test -v .

format: ${SOURCE}
	@gofmt -w ${SOURCE}

.PHONY: test format
