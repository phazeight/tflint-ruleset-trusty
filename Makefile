default: build

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	go build

.PHONY: install
install: build
	@mkdir -p ~/.tflint.d/plugins
	mv ./tflint-ruleset-trusty ~/.tflint.d/plugins/tflint-ruleset-trusty

.PHONY: snapshot
snapshot:
	goreleaser release --snapshot --rm-dist

.PHONY: release
release:
	@rm -rf ./dist
	GITHUB_TOKEN=$$( gh auth token ) goreleaser release
