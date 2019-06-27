export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
export VERSION=(unknown)
GO := go
ENV ?= dev
LDFLAGS ?= -X main.version=$(VERSION)
BUILDFLAGS ?= -a -ldflags '$(LDFLAGS)'
APPSOURCES := $(wildcard ./app/*.go storage/*/*.go activitypub/*.go internal/*/*.go)
PROJECT_NAME := $(shell basename $(PWD))

ifneq ($(ENV), dev)
	LDFLAGS += -s -w -extldflags "-static"
endif

ifeq ($(shell git describe --always > /dev/null 2>&1 ; echo $$?), 0)
export VERSION = $(shell git describe --always)
endif
ifeq ($(shell git describe --tags > /dev/null 2>&1 ; echo $$?), 0)
export VERSION = $(shell git describe --tags)
endif

BUILD := $(GO) build $(BUILDFLAGS)
TEST := $(GO) test $(BUILDFLAGS)

.PHONY: all run clean test coverage integration

all: app bootstrap

app: bin/app
bin/app: go.mod main.go $(APPSOURCES)
	$(BUILD) -tags $(ENV) -o $@ ./main.go

bootstrap: bin/bootstrap
bin/bootstrap: go.mod $(wildcard storage/*/*.go) cli/bootstrap/main.go
	$(BUILD) -tags $(ENV) -o $@ cli/bootstrap/main.go

run: app
	@./bin/app

clean:
	-$(RM) bin/*
	$(MAKE) -C tests $@


test: TEST_TARGET := ./{activitypub,app,storage,internal}/...
test:
	$(TEST) $(TEST_FLAGS) $(TEST_TARGET)

coverage: TEST_TARGET := .
coverage: TEST_FLAGS += -covermode=count -coverprofile $(PROJECT_NAME).coverprofile
coverage: test

integration:
	$(MAKE) -C tests $@
