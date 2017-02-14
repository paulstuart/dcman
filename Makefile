APP=dcman
BASE=/opt/dcman

export GOROOT := /usr/local/go
export GOPATH := /home/dcadmin/GO

#export PATH := bin:$(PATH)

export PATH := $(PATH):$(GOROOT)/bin:$(GOPATH)/bin

## simple makefile to log workflow
.PHONY: all test clean build install

GOFLAGS ?= $(GOFLAGS:)

all: install test

path:
	@echo $$PATH

build:
	@go build $(GOFLAGS) ./...

install: build
	#@go get -u $(GOFLAGS) ./...
	cp $(APP) $(BASE)
	cp -r assets $(BASE)
	cp -r sql $(BASE)

assets: FORCE
	cp -r assets $(BASE)
	cp -r sql $(BASE)

test: install
	@go test $(GOFLAGS) ./...

bench: install
	@go test -run=NONE -bench=. $(GOFLAGS) ./...

clean:
	@go clean $(GOFLAGS) -i ./...

FORCE:

## EOF
