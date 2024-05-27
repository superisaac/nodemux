GOFILES := $(shell find . -name '*.go')
GOFLAG :=
GOBUILD := CGO_ENABLED=0 go build -v

GOARCHS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64
buildarchdirs := $(foreach a,$(GOARCHS),build/arch/nodemux-$a)

build: bin/nodemux bin/nodemux-dail

all: test build

bin/nodemux: ${GOFILES}
	${GOBUILD} ${GOFLAG} -o $@ cmds/nodemux/main.go

bin/nodemux-dail: ${GOFILES}
	${GOBUILD} ${GOFLAG} -o $@ cmds/dail/main.go

test:
	go test -v ./...

clean:
	rm -rf build dist bin/nodemux bin/nodemux-dail

golint:
	go fmt ./...
	go vet ./...

run-example: bin/nodemux
	bin/nodemux -f examples/nodemux.example.yml -server examples/server.example.yml


# cross build distributions of multiple targets

archs:
	@for arch in $(GOARCHS); do \
		$(MAKE) dist/nodemux-$$arch.tar.gz; \
	done

dist/nodemux-%.tar.gz: build/arch/nodemux-%
	mkdir -p dist
	tar czvf $@ $<

build/arch/nodemux-%: ${GOFILES}
	GOOS=$(shell echo $@|cut -d- -f 2) \
	GOARCH=$(shell echo $@|cut -d- -f 3) \
	${GOBUILD} ${GOFLAG} -o $@/nodemux nodemux.go

.PHONY: build all test govet gofmt archs run-example
