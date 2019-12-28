BUILD_DIR = $(shell pwd)/build
RAGEL_VERSION = 6.9

export GOBIN = $(BUILD_DIR)/bin

RAGEL = $(GOBIN)/ragel
THRIFTRW = $(GOBIN)/thriftrw
GOLINT = $(GOBIN)/golint

LINT_EXCLUDES = \
	gen/internal/tests/ \
	idl/internal/lex.go \
	idl/internal/y.go \
	plugin/api/plugin.go \
	plugin/api/plugin_client.go \
	plugin/api/servicegenerator.go \
	plugin/api/servicegenerator_client.go \
	vendor/

# For tests on generated code, ignore deprecated warnings.
LINT_EXCLUDES += gen/.*_test.go:.*deprecated

# For main.go, ignore error string capitalized errors since these are
# user-facing errors.
LINT_EXCLUDES += main.go:.*error.*strings

# The "required" annotation is for easyjson.
LINT_EXCLUDES += unknown.*JSON.*option.*required

##############################################################################

RAGEL_TAR = $(BUILD_DIR)/src/ragel-$(RAGEL_VERSION).tar.gz

GO_FILES := $(shell \
	find . '(' -path '*/.*' -o -path './vendor' ')' -prune \
	-o -name '*.go' -print | cut -b3-)

# Installs dependencies listed in tools.go.
.PHONY: tools
tools:
	go list -json tools.go | jq -r '.Imports | .[]' | xargs -n1 go install

.PHONY: build
build: $(THRIFTRW)

.PHONY: ragel
ragel: $(RAGEL)

$(THRIFTRW):
	go install .

$(RAGEL_TAR):
	mkdir -p $(dir $@)
	curl -o $@ https://www.colm.net/files/ragel/ragel-$(RAGEL_VERSION).tar.gz

$(RAGEL): $(RAGEL_TAR)
	mkdir -p $(BUILD_DIR)/src/ragel
	tar -xz -C $(BUILD_DIR)/src/ragel --strip-components 1 < $(RAGEL_TAR)
	cd $(BUILD_DIR)/src/ragel && \
		./configure --prefix=$(BUILD_DIR) && \
		make install

.PHONY: generate
generate: tools $(RAGEL) $(THRIFTRW)
	PATH=$(GOBIN):$$PATH go generate ./...
	make -C ./gen/internal/tests
	./scripts/updateLicenses.sh

# Pipe lint output into this to filter out ignored portions.
LINT_FILTER := grep -v $(patsubst %,-e %, $(LINT_EXCLUDES))

.PHONY: lint
lint: $(GOLINT)
	@rm -rf lint.log
	@echo "Checking gofmt"
	@gofmt -e -s -l $(GO_FILES) 2>&1 | $(LINT_FILTER) | tee -a lint.log
	@echo "Checking govet"
	@go vet ./... 2>&1 | \
		grep -v '^#' | $(LINT_FILTER) | tee -a lint.log
	@echo "Checking golint"
	@$(GOLINT) ./... 2>&1 | $(LINT_FILTER) | tee -a lint.log
	@[ ! -s lint.log ]

.PHONY: verifyversion
verifyversion:
	$(eval CHANGELOG_VERSION := $(shell perl -ne '/^## \[(\S+?)\]/ && print "v$$1\n"' CHANGELOG.md | head -n1))
	$(eval INTHECODE_VERSION := $(shell perl -ne '/^const Version.*"([^"]+)".*$$/ && print "v$$1\n"' version/version.go))
	@if [ "$(INTHECODE_VERSION)" = "$(CHANGELOG_VERSION)" ]; then \
		echo "thriftrw-go: $(CHANGELOG_VERSION)"; \
	elif [ "$(CHANGELOG_VERSION)" = "vUnreleased" ]; then \
		echo "thriftrw-go (development): $(INTHECODE_VERSION)"; \
	else \
		echo "Version number in version/version.go does not match CHANGELOG.md"; \
		echo "version/version.go: $(INTHECODE_VERSION)"; \
		echo "CHANGELOG : $(CHANGELOG_VERSION)"; \
		exit 1; \
	fi

.PHONY: test
test: build verifyversion
	PATH=$(GOBIN):$$PATH go test -race ./...

# List of files we don't need to track coverage for.
# (Include a reason for each.)
#
# lex.go, y.go: Generated by Ragel and goyacc. Many cases cannot be exercised.
# mock_protocol.go: Generated by gomock, which is thoroughly tested.
# plugin/api/*.go: Generated by us. We're already testing codegen elsewhere.
COVER_IGNORE_FILES = \
	 idl/internal/lex.go \
	 idl/internal/y.go \
	 thrifttest/mock_protocol.go \
	 $(wildcard plugin/api/*.go)

# literal space
space :=
space +=

.PHONY: cover
cover:
	go test -v -covermode=atomic -coverprofile cover.full.out -coverpkg=./... ./...
	grep -v "$(subst $(space),\|,$(COVER_IGNORE_FILES))" cover.full.out > cover.out
	go tool cover -html=cover.out -o cover.html

.PHONY: clean
clean:
	go clean
