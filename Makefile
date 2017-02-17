.PHONY: all test build run
SHELL := /bin/bash

all: test build

test: build
	@./tests/run_integration_tests.sh

build:
	./node_modules/.bin/eslint .

run: build
	docker build -t who-is-who .
	@docker run -p 8080:80 --env-file=<(echo -e $(_ARKLOC_ENV_FILE)) who-is-who
