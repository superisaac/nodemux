
gofiles := $(shell find . -name '*.go')
goflag := -gcflags=-G=3

build: bin/nodeb

all: test build

bin/nodeb: ${gofiles}
	go build $(goflag) -o $@ nodeb.go

test:
	go test -v github.com/superisaac/nodeb/balancer
	go test -v github.com/superisaac/nodeb/chains
	go test -v github.com/superisaac/nodeb/server

clean:
	rm -rf build dist bin/nodeb

gofmt:
	go fmt balancer/*.go
	go fmt chains/*.go
	go fmt server/*.go
	go fmt nodeb.go

.PHONY: build all test gofmt dist
