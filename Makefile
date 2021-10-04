include node.mk
.DEFAULT_GOAL := test
NODE_VERSION := "v12"
$(eval $(call node-version-check,$(NODE_VERSION)))

.PHONY: all test lint lint-fix format format-all format-check run build-client
SHELL := /bin/bash
JS_FILES := $(shell find . -name "*.js" -not -path "./node_modules/*")
FORMATTED_FILES := $(JS_FILES)
MODIFIED_FORMATTED_FILES := $(shell git diff --name-only master $(FORMATTED_FILES))

PRETTIER := ./node_modules/.bin/prettier

all: test lint

build-client:
	$(MAKE) -C go-client

test: build-client lint
	@./tests/run_integration_tests.sh

lint: format-check
	./node_modules/.bin/eslint $(JS_FILES)

sync:
	node ./scripts/sync-users.js

format:
	@echo "Formatting modified files..."
	@$(PRETTIER) --write $(MODIFIED_FORMATTED_FILES)

format-all:
	@echo "Formatting all files..."
	@$(PRETTIER) --write $(FORMATTED_FILES)

format-check:
	@echo "Running format check..."
	@$(PRETTIER) --list-different $(FORMATTED_FILES) || \
		(echo -e "❌ \033[0;31m Prettier found discrepancies in the above files. Run 'make format' to fix.\033[0m" && false)

lint-fix:
	./node_modules/.bin/eslint --fix $(JS_FILES)

run: lint
	docker build -t who-is-who .
	@docker run -p 8081:80 --env-file=<(echo -e $(_ARKLOC_ENV_FILE)) \
		-v $(AWS_SHARED_CREDENTIALS_FILE):$(AWS_SHARED_CREDENTIALS_FILE) \
		who-is-who
