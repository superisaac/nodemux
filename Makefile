GOFILES := $(shell find . -name '*.go')
GOFLAG := -gcflags=-G=3
GOBUILD := GO111MODULE=on go build -v

build: bin/nodemux

all: test build

bin/nodemux: ${GOFILES}
	${GOBUILD} ${GOFLAG} -o $@ nodemux.go

test:
	go test -v ./...

clean:
	rm -rf build dist bin/nodemux

govet:
	go vet ./...

gofmt:
	go fmt ./...

.PHONY: build all test govet gofmt dist
