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

godeps: $(GOBIN)/godeps
	GOBIN=$(GOBIN) $(GOBIN)/godeps -u dependencies.tsv

$(GOBIN)/godeps: $(GOBIN)
	GOBIN=$(GOBIN) go get github.com/rogpeppe/godeps

ifeq ($(MAKE_GODEPS),true)
.PHONY: deps
deps: godeps
else
deps:
	@echo "Skipping godeps. export MAKE_GODEPS = true to enable."
endif

build: deps
	GOBIN=$(GOBIN) go build -a $(PROJECT)/...

install: deps
	GOBIN=$(GOBIN) go install -v $(PROJECT)/...

check: build
	GOBIN=$(GOBIN) go test -test.timeout=1200s $(PROJECT)/...

land: check race

race: build
	GOBIN=$(GOBIN) go test -test.timeout=1200s -race $(PROJECT)/...

.PHONY: build check install clean race land 
