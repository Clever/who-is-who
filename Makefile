include golang.mk
.DEFAULT_GOAL := test # override default goal set in library makefile

SHELL := /bin/bash
PKG = github.com/Clever/who-is-who
PKGS := $(shell go list ./... | grep -v /vendor)
EXECUTABLE := who-is-who
.PHONY: test build vendor run

$(eval $(call golang-version-check,1.5))

build:
	go build -o bin/$(EXECUTABLE) $(PKG)

test: $(PKGS)

$(PKGS): golang-test-all-deps
	$(call golang-fmt,$@)
	$(call golang-lint,$@)
	$(call golang-vet,$@)
	./integration_test.sh $@

vendor: golang-godep-vendor-deps
	$(call golang-godep-vendor,$(PKGS))

run: build
	bin/$(EXECUTABLE)
