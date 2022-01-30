GOFILES := $(shell find . -name '*.go')
GOFLAG := -gcflags=-G=3
GOBUILD := GO111MODULE=on go build -v

build: bin/nodemux

all: test build

bin/nodemux: ${GOFILES}
	${GOBUILD} ${GOFLAG} -o $@ nodemux.go

test:
	go test ${GOFLAG} -v github.com/superisaac/nodemux/core
	go test ${GOFLAG} -v github.com/superisaac/nodemux/chains
	go test ${GOFLAG} -v github.com/superisaac/nodemux/server

clean:
	rm -rf build dist bin/nodemux

gofmt:
	go fmt core/*.go
	go fmt chains/*.go
	go fmt server/*.go
	go fmt nodemux.go

.PHONY: build all test gofmt dist
