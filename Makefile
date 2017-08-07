##=======================================================================##
## Makefile
## Created: Wed Aug 05 14:35:14 PDT 2015 @941 /Internet Time/
# :mode=makefile:tabSize=3:indentSize=3:
## Purpose:
##======================================================================##

SHELL=/bin/bash
PROJECT_NAME = SchemaBuilder
GPATH = $(shell pwd)

.PHONY: fmt test install build scrape clean

install: fmt test
	@GOPATH=${GPATH} go build -o ${PROJECT_NAME} main.go

build: fmt test
	@GOPATH=${GPATH} go build -o ${PROJECT_NAME} main.go

fmt:
	@GOPATH=${GPATH} gofmt -s -w *.go
	@GOPATH=${GPATH} gofmt -s -w schema_builder

test: fmt
	@GOPATH=${GPATH} go test -v -bench=. -test.benchmem

scrape:
	@find src -type d -name '.hg' -or -type d -name '.git' | xargs rm -rf

clean:
	@GOPATH=${GPATH} go clean
