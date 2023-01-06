GOFILES := $(shell find . -name '*.go')
GOFLAG :=
GOBUILD := GO111MODULE=on go build -v

build: bin/nodemux

all: test build

bin/nodemux: ${GOFILES}
	${GOBUILD} ${GOFLAG} -o $@ nodemux.go

test:
	go test -v ./...

clean:
	rm -rf build dist bin/nodemux

golint:
	go fmt ./...
	go vet ./...

run-example: bin/nodemux
	bin/nodemux -f examples/nodemux.example.yml -server examples/server.example.yml

.PHONY: build all test govet gofmt dist run-example
