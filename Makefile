default: build

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build

PLUGIN_VERSION := $(shell grep 'Version\s*=' project/main.go | sed 's/.*"\(.*\)".*/\1/')
PLUGIN_DIR     := $(HOME)/.tflint.d/plugins/github.com/phazeight/tflint-ruleset-trusty/$(PLUGIN_VERSION)

.PHONY: install
install: build
	@mkdir -p $(PLUGIN_DIR)
	mv ./tflint-ruleset-trusty $(PLUGIN_DIR)/tflint-ruleset-trusty

.PHONY: snapshot
snapshot:
	goreleaser release --snapshot --rm-dist

.PHONY: release
release:
	@rm -rf ./dist
	GITHUB_TOKEN=$$( gh auth token ) goreleaser release
