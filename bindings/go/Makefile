# Makefile for the go bindings of notmuch

export GOPATH      ?= $(shell pwd)
export CGO_CFLAGS  ?= -I../../../../lib
export CGO_LDFLAGS ?= -L../../../../lib

GO         ?= go
GOFMT      ?= gofmt

all: notmuch notmuch-addrlookup notmuch-mtag

.PHONY: notmuch
notmuch:
	$(GO) install notmuch

.PHONY: goconfig
goconfig:
	if [ ! -d src/github.com/kless/goconfig/config ]; then \
	    $(GO) get github.com/kless/goconfig/config; \
	fi

.PHONY: notmuch-addrlookup
notmuch-addrlookup: notmuch goconfig
	$(GO) install notmuch-addrlookup

.PHONY: notmuch-mtag
notmuch-mtag: notmuch goconfig
	$(GO) install notmuch-mtag

.PHONY: format
format:
	$(GOFMT) -w=true $(GOFMT_OPTS) src/notmuch
	$(GOFMT) -w=true $(GOFMT_OPTS) src/notmuch-addrlookup
	$(GOFMT) -w=true $(GOFMT_OPTS) src/notmuch-mtag

.PHONY: check-format
check-format:
	$(GOFMT) -d=true $(GOFMT_OPTS) src/notmuch
	$(GOFMT) -d=true $(GOFMT_OPTS) src/notmuch-addrlookup
	$(GOFMT) -d=true $(GOFMT_OPTS) src/notmuch-mtag

.PHONY: clean
clean:
	$(GO) clean notmuch
	$(GO) clean notmuch-addrlookup
	$(GO) clean notmuch-mtag
	rm -rf pkg bin
