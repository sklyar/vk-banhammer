V := @

VERSION = $(shell git describe --match "v[0-9]*" --abbrev=0 --tags)
COMMIT = $(shell git rev-parse --short HEAD)
DATE = $(shell date +%F_%H-%M-%S)

NAME = vk-banhammer
VCS = github.com
ORG = sklyar

.PHONY: install
install:
	$(call print-target)
	$(V)go mod tidy

.PHONY: lint
lint:
	$(call print-target)
	$(V)golangci-lint run

.PHONY: test
test: GO_TEST_FLAGS += -race
test:
	$(call print-target)
	$(V)go test $(GO_TEST_FLAGS) --tags=$(GO_TEST_TAGS) ./...

.PHONY: clean
clean:
	$(call print-target)
	$(V)golangci-lint cache clean

.PHONY: build
build:
	$(call print-target)
	$(V)go build -mod=vendor -o $(NAME) ./cmd/$(NAME) -ldflags "-X 'main.version=${VERSION}'"

.PHONY: docker-build
docker-build:
	$(call print-target)
	$(call check_defined, VERSION)
	$(V)docker build -t sklyar/banhammer:${VERSION} --build-arg VERSION=${VERSION} .

.PHONY: docker-build-local
docker-build-local: ## build in docker for local env
	$(V)docker build -t sklyar/banhammer:local .

.PHONY: diff
diff: ## git diff
	$(call print-target)
	git diff --exit-code
	RES=$$(git status --porcelain) ; if [ -n "$$RES" ]; then echo $$RES && exit 1 ; fi

define print-target
    @printf "Executing target: \033[36m$@\033[0m\n"
endef
