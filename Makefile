.PHONY: all test lint format run build-client
SHELL := /bin/bash
JS_FILES := $(shell find . -name "*.js" -not -path "./node_modules/*")

all: test lint

build-client:
	go get -d -t github.com/Clever/who-is-who/go-client # fetch deps
	go build go-client/client.go # ensures go client compiles
	go test github.com/Clever/who-is-who/go-client # run tests

test: build-client lint
	@./tests/run_integration_tests.sh

lint:
	./node_modules/.bin/eslint $(JS_FILES)

sync:
	node ./scripts/sync-users.js

format:
	./node_modules/.bin/prettier --bracket-spacing false --write $(JS_FILES)
	./node_modules/.bin/eslint --fix $(JS_FILES)

run: lint
	docker build -t who-is-who .
	@docker run -p 8081:80 --env-file=<(echo -e $(_ARKLOC_ENV_FILE)) who-is-who
