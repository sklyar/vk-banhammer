# suppress output, run `make XXX V=` to be verbose
V := @

# Common flags
NAME = vk-banhammer
VCS = github.com
ORG = sklyar

.PHONY: lint
lint:
	$(V)golangci-lint run

.PHONY: test
test: GO_TEST_FLAGS += -race
test:
	$(V)go test -mod=vendor $(GO_TEST_FLAGS) --tags=$(GO_TEST_TAGS) ./...

.PHONY: generate
generate:
	$(V)go generate -x ./...

.PHONY: clean
clean:
	$(V)golangci-lint cache clean

.PHONY: vendor
vendor:
	$(V)GOPRIVATE=${VCS}/* go mod tidy
	$(V)GOPRIVATE=${VCS}/* go mod vendor
	$(V)git add vendor go.mod go.sum
