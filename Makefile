# Copyright 2023 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

MKDIR    ?= mkdir
TR       ?= tr
DIST_DIR ?= $(CURDIR)/dist
HELM     ?= go tool -modfile $(CURDIR)/hack/tools/go.mod helm
GO       ?= go
GOLANGCI_LINT ?= go tool -modfile $(CURDIR)/hack/tools/golangci-lint/go.mod golangci-lint
GOLANGCI_LINT_CONFIG ?= .golangci.yaml
GO_PACKAGES ?= ./...
GO_BUILD_PACKAGES ?= ./...
GO_COVERAGE_FILE ?= cover.out
CHART_PATH ?= deployments/helm/enflame-dra-driver
HELM_RELEASE_NAME ?= enflame-dra-driver
# Chart validation rejects the default namespace unless allowDefaultNamespace=true.
HELM_TEMPLATE_ARGS ?= --namespace kube-system

export IMAGE_GIT_TAG ?= $(shell git describe --tags --always --dirty --match 'v*')
export CHART_GIT_TAG ?= $(shell git describe --tags --always --dirty --match 'chart/*')

include $(CURDIR)/common.mk

CMDS := $(patsubst ./cmd/%/,%,$(sort $(dir $(wildcard ./cmd/*/))))
CMD_TARGETS := $(patsubst %,cmd-%, $(CMDS))

CHECK_TARGETS := assert-fmt vet lint ineffassign misspell
MAKE_TARGETS := binaries build check fmt fmt-check test test-ci examples cmds coverage generate ci verify-build verify-helm-chart $(CHECK_TARGETS)

TARGETS := $(MAKE_TARGETS) $(CMD_TARGETS)

.PHONY: $(TARGETS)

GOOS ?= linux

binaries: cmds
ifneq ($(PREFIX),)
cmd-%: COMMAND_BUILD_OPTIONS = -o $(PREFIX)/$(*)
endif
cmds: $(CMD_TARGETS)
# Cross-compiling to linux from macOS disables cgo by default. Production
# images must build on linux/amd64 with CGO_ENABLED=1 and libefml present.
# GitHub Actions / make ci use CGO_ENABLED=0 and the EFML stub.
CGO_ENABLED ?= $(shell if [ "$(GOOS)" = "linux" ] && [ "$$(uname -s)" = "Linux" ]; then echo 1; else echo 0; fi)

$(CMD_TARGETS): cmd-%:
	CGO_ENABLED=$(CGO_ENABLED) CGO_LDFLAGS_ALLOW='-Wl,--unresolved-symbols=ignore-in-object-files' GOOS=$(GOOS) \
		$(GO) build -ldflags "-s -w -X main.version=$(VERSION)" $(COMMAND_BUILD_OPTIONS) $(MODULE)/cmd/$(*)

build:
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) $(GO) build ./...

examples: $(EXAMPLE_TARGETS)
$(EXAMPLE_TARGETS): example-%:
	GOOS=$(GOOS) $(GO) build ./examples/$(*)

all: check test build binary
check: $(CHECK_TARGETS)

# dynamia-ai Go CI contract: non-mutating checks that do not require host libefml.
# Race detection needs cgo; CI therefore skips -race and links the EFML stub.
ci: fmt-check vet lint test-ci verify-build

fmt:
	$(GOLANGCI_LINT) fmt --config "$(GOLANGCI_LINT_CONFIG)"

fmt-check:
	$(GOLANGCI_LINT) fmt --diff --config "$(GOLANGCI_LINT_CONFIG)"

assert-fmt: fmt-check

ineffassign:
	go tool -modfile hack/tools/go.mod ineffassign $(MODULE)/...

lint:
	$(GOLANGCI_LINT) run --config "$(GOLANGCI_LINT_CONFIG)" $(GO_PACKAGES)

misspell:
	go tool -modfile hack/tools/go.mod misspell $(MODULE)/...

vet:
	CGO_ENABLED=0 $(GO) vet $(GO_PACKAGES)

# Ensure that all log calls support contextual logging.
.PHONY: logcheck
logcheck:
	go tool -modfile hack/tools/go.mod logcheck -check-contextual -check-deprecations ./...

COVERAGE_FILE := $(GO_COVERAGE_FILE)
test: logcheck
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) $(GO) test -v -coverprofile=$(COVERAGE_FILE) $(GO_PACKAGES)

# CI entry used by make ci / Go CI workflow (always stub EFML; no -race).
test-ci:
	CGO_ENABLED=0 $(GO) test -covermode=atomic -coverprofile="$(GO_COVERAGE_FILE)" $(GO_PACKAGES)
	test -s "$(GO_COVERAGE_FILE)"

verify-build:
	CGO_ENABLED=0 GOOS=$(GOOS) $(GO) build $(GO_BUILD_PACKAGES)

# Helm CI contract: lint + render. Chart does not regenerate README/schema in CI.
verify-helm-chart:
	$(HELM) lint "$(CHART_PATH)"
	$(HELM) template \
		"$(HELM_RELEASE_NAME)" \
		"$(CHART_PATH)" \
		$(HELM_TEMPLATE_ARGS) >/dev/null

coverage: test
	cat $(COVERAGE_FILE) | grep -v "_mock.go" > $(COVERAGE_FILE).no-mocks
	$(GO) tool cover -func=$(COVERAGE_FILE).no-mocks

generate: generate-deepcopy generate-conversion

generate-deepcopy:
	for api in $(APIS); do \
		rm -f $${api}/zz_generated.deepcopy.go; \
		go tool -modfile hack/tools/go.mod controller-gen \
			object:headerFile=$(CURDIR)/hack/boilerplate.generatego.txt \
			paths=$${api}/ \
			output:object:dir=$${api}; \
	done

generate-conversion:
	for api in $(APIS); do \
		rm -f $${api}/zz_generated.conversion.go; \
		go tool -modfile hack/tools/go.mod conversion-gen \
			--go-header-file=$(CURDIR)/hack/boilerplate.generatego.txt \
			--output-file=zz_generated.conversion.go \
			$${api}/; \
	done

.PHONY: push-release-artifacts
push-release-artifacts:
	CHART_VERSION="$${CHART_GIT_TAG##chart/}" \
		HELM=$(HELM) \
		demo/scripts/push-driver-chart.sh
	export DRIVER_IMAGE_TAG="${IMAGE_GIT_TAG}"; \
	demo/scripts/push-driver-image.sh
