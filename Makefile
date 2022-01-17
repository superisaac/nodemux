
gofiles := $(shell find . -name '*.go')
goflag := -gcflags=-G=3

build: bin/nodepool

all: test build

bin/nodepool: ${gofiles}
	go build ${goflag} -o $@ nodepool.go

test:
	go test ${goflag} -v github.com/superisaac/nodepool/cfg
	go test ${goflag} -v github.com/superisaac/nodepool/balancer
	go test ${goflag} -v github.com/superisaac/nodepool/chains
	go test ${goflag} -v github.com/superisaac/nodepool/server

clean:
	rm -rf build dist bin/nodepool

gofmt:
	go fmt cfg/*.go
	go fmt balancer/*.go
	go fmt chains/*.go
	go fmt server/*.go
	go fmt nodepool.go

.PHONY: build all test gofmt dist
