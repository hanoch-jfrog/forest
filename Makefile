SHELL := /bin/bash

.DEFAULT_GOAL = help

GOCMD ?= go
TEST_TAGS ?= -tags=test
.DEFAULT_GOAL = build

help:				## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

# BUILD:

build:	clean			## Build Forest plugin
	$(GOCMD) build

fmt-fix:			## Gofmt fix errors
	gofmt -w -s .

vet:				## GoVet
	$(GOCMD) vet $(TEST_TAGS) ./...

clean:				## Clean from created bins
	rm -f forest main jfrog-cli-plugin-template main

run:			## Run the plugin
	$(GOCMD) run main.go


# TEST EXECUTION

test:				## Run all tests
	export GOMAXPROCS=4
	time $(GOCMD) test ./... $(TEST_TAGS) -count=1

test-list: 			## List all tests
	$(GOCMD) list ./...


# PLUGIN INSTALLATION

install:			## Install the plugin to jfrog cli
	jfrog plugin install hello-frog

uninstall:			## Uninstall the plugin to jfrog cli
	jfrog plugin uninstall hello-frog

.PHONY: help build run clean vet test test-list fmt-fix install uninstall
