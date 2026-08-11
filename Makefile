export GOPATH?=$(shell go env GOPATH)
export GOPROXY=https://proxy.golang.org
export GO111MODULE=on

HOSTNAME=hashicorp.com
NAMESPACE=anexia-it
NAME=anxcloud
BINARY=terraform-provider-${NAME}
VERSION=0.11.1
OS_ARCH?=$(shell go env GOOS)_$(shell go env GOARCH)
GOLDFLAGS= -s -w -X main.version=$(VERSION)

# Common deployment platforms. Keep this aligned with .goreleaser.yml.
PLATFORMS := \
	darwin_amd64 \
	darwin_arm64 \
	linux_amd64 \
	linux_arm64 \
	windows_amd64 \
	windows_arm64

TARGET_OS?=$(shell go env GOOS)
TARGET_ARCH?=$(shell go env GOARCH)
TARGET_ARM?=6
TARGET_EXT=$(if $(filter windows,$(TARGET_OS)),.exe,)
TARGET_BINARY=./bin/${BINARY}_${VERSION}_${TARGET_OS}_${TARGET_ARCH}${TARGET_EXT}

GOFMT_FILES  := $(shell find ./anxcloud -name '*.go' |grep -v vendor)

default: install

.PHONY: build
build: fmtcheck go-lint
	go build -o ${BINARY}

.PHONY: release
release: fmtcheck lint test testacc cross-build

.PHONY: build-target
build-target:
	@mkdir -p ./bin
	@echo "==> Building ${TARGET_OS}/${TARGET_ARCH}..."
	CGO_ENABLED=0 GOOS=${TARGET_OS} GOARCH=${TARGET_ARCH} GOARM=${TARGET_ARM} \
		go build -trimpath -o "${TARGET_BINARY}" -ldflags "$(GOLDFLAGS)"

.PHONY: cross-build
cross-build:
	@set -e; for platform in $(PLATFORMS); do \
		target_os=$${platform%_*}; \
		target_arch=$${platform#*_}; \
		$(MAKE) --no-print-directory build-target TARGET_OS=$$target_os TARGET_ARCH=$$target_arch; \
	done

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

.PHONY: test
test: fmtcheck
	go test ./... -v -coverprofile coverage.out $(TESTARGS)
	go tool cover -html=coverage.out -o coverage.html

.PHONY: testacc
testacc: fmtcheck
	@if [ "$(TESTARGS)" = "-run=TestAccXXX" ]; then \
		echo ""; \
		echo "Error: Skipping example acceptance testing pattern. Update TESTARGS to match the test naming in the relevant *_test.go file."; \
		echo "Example:"; \
		echo ""; \
		echo "    make testacc TESTARGS='-run=TestAccAnxCloudVirtualServer'"; \
		echo ""; \
		exit 1; \
	fi
	TF_ACC=1 go test ./... -v -coverprofile coverage.out $(TESTARGS) -timeout 120m
	go tool cover -html=coverage.out -o coverage.html

.PHONY: fmt
fmt:
	gofmt -w $(GOFMT_FILES)

.PHONY: fmtcheck
fmtcheck:
	@./scripts/gofmtcheck.sh

.PHONY: depscheck
depscheck:
	@echo "==> Checking source code dependencies..."
	@go mod tidy
	@git diff --exit-code -- go.mod go.sum || \
		(echo; echo "Found differences in go.mod/go.sum files. Run 'go mod tidy' or revert go.mod/go.sum changes."; exit 1)

.PHONY: tools/misspell
tools/misspell:
	cd tools && go build -o . github.com/client9/misspell/cmd/misspell

.PHONY: tools/golangci-lint
tools/golangci-lint:
	cd tools && go build -o . github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: tools/terrafmt
tools/terrafmt:
	cd tools && go build -o . github.com/katbyte/terrafmt

.PHONY: tools/tfplugindocs
tools/tfplugindocs:
	cd tools && go build -o . github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

.PHONY: docs-generate
docs-generate: tools/tfplugindocs
	@echo -n "This operation will clear the /docs directory. Changes will be lost! Do you want to proceed? [y/N] " && read ans && [ $${ans:-N} = y ]
	@tools/tfplugindocs
	@echo "You need to manually patch some markdown files in /docs!"
# https://github.com/hashicorp/terraform-plugin-docs/issues/28#issuecomment-768299611

.PHONY: docs-lint
docs-lint: misspell terrafmt

.PHONY: docs-lint-fix
docs-lint-fix: tools/misspell tools/terrafmt
	@echo "==> Applying automatic docs linter fixes..."
	@tools/misspell -w -source=text docs/
	@tools/terrafmt fmt ./docs --pattern '*.md'

.PHONY: go-lint
go-lint: tools/golangci-lint
	@echo "==> Checking source code against linters..."
	@tools/golangci-lint run ./...

.PHONY: terrafmt
terrafmt: tools/terrafmt
	@tools/terrafmt diff ./docs --check --pattern '*.md' --quiet || (echo; \
		echo "Unexpected differences in docs HCL formatting."; \
		echo "To see the full differences, run: tools/terrafmt diff ./docs --pattern '*.md'"; \
		echo "To automatically fix the formatting, run 'make docs-lint-fix' and commit the changes."; \
		exit 1)

.PHONY: misspell
misspell: tools/misspell
	@tools/misspell -error -source=text docs/ || (echo; \
		echo "Unexpected misspelling found in docs files."; \
		echo "To automatically fix the misspelling, run 'make docs-lint-fix' and commit the changes."; \
		exit 1)

.PHONY: lint
lint: go-lint docs-lint
