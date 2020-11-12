HOSTNAME=hashicorp.com
NAMESPACE=anexia-it
NAME=anxcloud
BINARY=terraform-provider-${NAME}
VERSION=0.1
OS_ARCH=linux_amd64

TEST?=$$(go list ./... | grep -v 'vendor')
GOFMT_FILES  := $$(find $(PROVIDER_DIR) -name '*.go' |grep -v vendor)

default: install

.PHONY: build
build:
	go build -o ${BINARY}

.PHONY: release
release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_windows_amd64

.PHONY: install
install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

.PHONY: test
test: fmtcheck
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

.PHONY: testacc
testacc: fmtcheck
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

.PHONY: fmt
fmt:
	gofmt -w $(GOFMT_FILES)

.PHONY: fmtcheck
fmtcheck:
	@./scripts/gofmtcheck.sh

.PHONY: tools
tools:
	go install github.com/bflad/tfproviderdocs
	go install github.com/client9/misspell/cmd/misspell
	go install github.com/katbyte/terrafmt
	go mod tidy
	go mod vendor

.PHONY: docs-lint
docs-lint: tools
	@echo "==> Checking docs against linters..."
	@misspell -error -source=text ./docs || (echo; \
		echo "Unexpected mispelling found in docs files."; \
		echo "To automatically fix the misspelling, run 'make docs-lint-fix' and commit the changes."; \
		exit 1)
	@echo "==> Running markdownlint-cli using DOCKER='$(DOCKER)', DOCKER_RUN_OPTS='$(DOCKER_RUN_OPTS)' and DOCKER_VOLUME_OPTS='$(DOCKER_VOLUME_OPTS)'"
	@$(DOCKER) run $(DOCKER_RUN_OPTS) -v $(PROVIDER_DIR):/workspace:$(DOCKER_VOLUME_OPTS) -w /workspace 06kellyjac/markdownlint-cli ./docs || (echo; \
		echo "Unexpected issues found in docs Markdown files."; \
		echo "To apply any automatic fixes, run 'make docs-lint-fix' and commit the changes."; \
		exit 1)
	@echo "==> Running terrafmt diff..."
	@terrafmt diff ./docs --check --pattern '*.markdown' --quiet || (echo; \
		echo "Unexpected differences in docs HCL formatting."; \
		echo "To see the full differences, run: terrafmt diff ./docs --pattern '*.markdown'"; \
		echo "To automatically fix the formatting, run 'make docs-lint-fix' and commit the changes."; \
		exit 1)
	@echo "==> Statically compiling provider for tfproviderdocs..."
	@env CGO_ENABLED=0 GOOS=$$(go env GOOS) GOARCH=$$(go env GOARCH) go build -a -o $(TF_PROV_DOCS)/terraform-provider-kubernetes
	@echo "==> Getting provider schema for tfproviderdocs..."
		@$(DOCKER) run $(DOCKER_RUN_OPTS) -v $(TF_PROV_DOCS):/workspace:$(DOCKER_VOLUME_OPTS) -w /workspace hashicorp/terraform:0.12.29 init
		@$(DOCKER) run $(DOCKER_RUN_OPTS) -v $(TF_PROV_DOCS):/workspace:$(DOCKER_VOLUME_OPTS) -w /workspace hashicorp/terraform:0.12.29 providers schema -json > $(TF_PROV_DOCS)/schema.json
	@echo "==> Running tfproviderdocs..."
	@tfproviderdocs check -providers-schema-json $(TF_PROV_DOCS)/schema.json -provider-name kubernetes
	@rm -f $(TF_PROV_DOCS)/schema.json $(TF_PROV_DOCS)/terraform-provider-kubernetes

.PHONY: docs-lint-fix
docs-lint-fix: tools
	@echo "==> Applying automatic docs linter fixes..."
	@misspell -w -source=text ./docs
	@echo "==> Running markdownlint-cli --fix using DOCKER='$(DOCKER)', DOCKER_RUN_OPTS='$(DOCKER_RUN_OPTS)' and DOCKER_VOLUME_OPTS='$(DOCKER_VOLUME_OPTS)'"
	@$(DOCKER) run $(DOCKER_RUN_OPTS) -v $(PROVIDER_DIR):/workspace:$(DOCKER_VOLUME_OPTS) -w /workspace 06kellyjac/markdownlint-cli --fix ./docs
	@echo "==> Fixing docs terraform blocks code with terrafmt..."
	@terrafmt fmt ./docs --pattern '*.markdown'
