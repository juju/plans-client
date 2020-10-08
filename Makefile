# Copyright 2017 Canonical Ltd.
# # Licensed under the GPLv3, see LICENCE file for details.
#
PROJECT := github.com/juju/plans-client

ifndef GOBIN
GOBIN := $(shell mkdir -p $(GOPATH)/bin; realpath $(GOPATH))/bin
else
REAL_GOBIN := $(shell mkdir -p $(GOBIN); realpath $(GOBIN))
GOBIN := $(REAL_GOBIN)
endif

build:
	GOBIN=$(GOBIN) go build -a $(PROJECT)/...

install:
	GOBIN=$(GOBIN) go install -v $(PROJECT)/...

check: build
	GOBIN=$(GOBIN) go test -test.timeout=1200s $(PROJECT)/...

land: check race

race: build
	GOBIN=$(GOBIN) go test -test.timeout=1200s -race $(PROJECT)/...

.PHONY: build check install clean race land 
