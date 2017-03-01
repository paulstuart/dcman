APP=dcman
BASE=/opt/dcman

SHELL := /bin/bash

export GOROOT := /usr/local/go
export GOPATH := /home/dcadmin/GO

#export PATH := bin:$(PATH)

export PATH := $(PATH):$(GOROOT)/bin:$(GOPATH)/bin

GOFLAGS ?= $(GOFLAGS:)

all: install test

path:
	@echo $$PATH

build:
	@go build $(GOFLAGS) ./...

# for building static distribution on Alpine Linux
# https://dominik.honnef.co/posts/2015/06/go-musl/#flavor-be-gone
compile:
	CC=/usr/bin/x86_64-alpine-linux-musl-gcc go build --ldflags '-linkmode external -extldflags "-static"'

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

get:
	go get -u

archive:
	tar -czf /shared/dcman.tgz dcman *config* assets data*

copy:
	cp -r /go/bin/* /shared/
	cp -r assets /shared
	cp -r *cfg /shared

fresh: get build

latest: fresh archive

pkg: compile archive

.PHONY: all test clean build compile install copy fresh get latest pkg

## EOF
