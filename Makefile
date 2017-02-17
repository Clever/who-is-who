.PHONY: all test lint format run
SHELL := /bin/bash
JS_FILES := $(shell find . -name "*.js" -not -path "./node_modules/*")

all: test lint

test: lint
	@./tests/run_integration_tests.sh

lint:
	./node_modules/.bin/eslint $(JS_FILES)

format:
	./node_modules/.bin/prettier --bracket-spacing false --single-quote --write $(JS_FILES)
	./node_modules/.bin/eslint --fix $(JS_FILES)

run: lint
	docker build -t who-is-who .
	@docker run -p 8080:80 --env-file=<(echo -e $(_ARKLOC_ENV_FILE)) who-is-who
