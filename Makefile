SHELL := /bin/bash
PKG = github.com/Clever/who-is-who
PKGS = $(PKG)

.PHONY: test golint README.md README

golint:
	@go get github.com/golang/lint/golint

test:
	go get ./...
	./integration_tests.sh

$(PKGS): golint README
	@go get -d -t $@
	@gofmt -w=true $(GOPATH)/src/$@*/**.go
ifneq ($(NOLINT),1)
	@echo "LINTING..."
	@PATH=$(PATH):$(GOPATH)/bin golint $(GOPATH)/src/$@*/**.go
	@echo ""
endif
ifeq ($(COVERAGE),1)
	@go test -cover -coverprofile=$(GOPATH)/src/$@/c.out $@ -test.v
	@go tool cover -html=$(GOPATH)/src/$@/c.out
else
	@echo "TESTING..."
	@go test $@ -test.v
endif

run:
	@go build
	./who-is-who
