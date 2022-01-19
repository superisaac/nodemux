
gofiles := $(shell find . -name '*.go')
goflag := -gcflags=-G=3

build: bin/nodemux

all: test build

bin/nodemux: ${gofiles}
	go build ${goflag} -o $@ nodemux.go

test:
	go test ${goflag} -v github.com/superisaac/nodemux/balancer
	go test ${goflag} -v github.com/superisaac/nodemux/chains
	go test ${goflag} -v github.com/superisaac/nodemux/server

clean:
	rm -rf build dist bin/nodemux

gofmt:
	go fmt balancer/*.go
	go fmt chains/*.go
	go fmt server/*.go
	go fmt nodemux.go

.PHONY: build all test gofmt dist
