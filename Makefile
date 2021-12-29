
gofiles := $(shell find . -name '*.go')
goflag := -gcflags=-G=3

build:

all: test build

test:
	go test -v github.com/superisaac/nodeb/balancer
	go test -v github.com/superisaac/nodeb/chains

clean:
	rm -rf build dist

gofmt:
	go fmt balancer/*.go
	go fmt chains/*.go

.PHONY: build all test gofmt dist
