
gofiles := $(shell find . -name '*.go')
goflag := -gcflags=-G=3

build:

all: test build

test:
	go test -v github.com/superisaac/nodeb/balancer

clean:
	rm -rf build dist

gofmt:
	go fmt balancer/*.go

.PHONY: build all test gofmt dist
