.PHONY: all test build run
SHELL := /bin/bash
JS_FILES := $(shell find . -name "*.js" -not -path "./node_modules/*")

all: test build

test: build
	@./tests/run_integration_tests.sh

build:
	./node_modules/.bin/eslint .

lint:
	./node_modules/.bin/eslint $(JS_FILES)

format:
	./node_modules/.bin/prettier --bracket-spacing false --single-quote --write $(JS_FILES)
	./node_modules/.bin/eslint --fix $(JS_FILES)

run: build
	docker build -t who-is-who .
	@docker run -p 8080:80 --env-file=<(echo -e $(_ARKLOC_ENV_FILE)) who-is-who
