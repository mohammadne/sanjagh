# Image URL to use all building/pushing image targets
IMAGE ?= ghcr.io/mohammadne/sanjagh

# Set number of cores to use in integration tests parallel execution
PROCS ?= $(shell grep -c 'cpu[0-9]' /proc/stat)

COMMIT ?= $(shell git rev-parse HEAD)
VERSION ?= $(shell git describe --tags --always --match=v*)
BUILD_TIME := $(shell LANG=en_US date)
IMAGE_ARGS := --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT)
LD_FLAGS := -X 'git.cafebazaar.ir/cloud/openstack/engine/neumann/cmd.BuildVersion=$(VERSION)' \
            -X 'git.cafebazaar.ir/cloud/openstack/engine/neumann/cmd.BuildCommit=$(COMMIT)' \
            -X 'git.cafebazaar.ir/cloud/openstack/engine/neumann/cmd.BuildTime=$(BUILD_TIME)'
GO_ENV   ?= CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOPRIVATE=git.cafebazaar.ir GOFLAGS=-mod=vendor
GO_FLAGS := $(GO_ENV) go build -ldflags="$(LD_FLAGS)"
GO_BUILD := $(if $(BUILD_TAGS),$(GO_FLAGS) -tags $(BUILD_TAGS),$(GO_FLAGS))

# Set the Operator SDK version to use. By default, what is installed on the system is used.
# This is useful for CI or a project to utilize a specific version of the operator-sdk toolkit.
OPERATOR_SDK_VERSION ?= v1.31.0

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.26.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: controller-gen ## Generate CustomResourceDefinition objects.
	$(CONTROLLER_GEN) crd paths="./api/..." output:crd:artifacts:config=deployment/sanjagh/crds

.PHONY: generate
generate: controller-gen ## Generate apis code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

.PHONY: functional-test
functional-test: fmt vet ## Run functional tests.
	mkdir -p coverage_results
	ginkgo run -v -p --trace --procs 5 --poll-progress-after 300s -cover -coverpkg="../../../controllers/...,../../../pkg/...,../../../projectconfig/..." --keep-separate-coverprofiles --coverprofile=../../../coverage_results/functional-coverage.out tests/suites/functional

.PHONY: integration-test
integration-test: fmt vet ## Run integration tests.
	mkdir -p coverage_results
	ginkgo run -p --procs 10 --poll-progress-after 300s -cover -coverpkg="../../../controllers/...,../../../pkg/...,../../../projectconfig/..." --keep-separate-coverprofiles --coverprofile=../../../coverage_results/integration-coverage.out tests/suites/integration

.PHONY: unit-test
unit-test: fmt vet ## Run unit tests.
	mkdir -p coverage_results
	$(GO_ENV) go test -cover -coverpkg="./controllers/...,./pkg/...,./projectconfig/..." -coverprofile=./coverage_results/unit-coverage.out `go list ./... | grep -Ev "integration|functional"`

##@ Build

.PHONY: pre-process
pre-process: manifests generate fmt vet

.PHONY: build
build: pre-process ## Build sangagh binary.
	go build -o bin/manager main.go

.PHONY: run
run: pre-process ## Run sangagh from your host.
	go run ./main.go

.PHONY: docker-build
docker-build: test ## Build docker image with the manager.
	docker build -t ${IMAGE}:${VERSION} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMAGE}:${VERSION}

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: version
version:
	@echo $(VERSION)

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	kubectl apply -f deployment/sangagh/crds

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	kubectl delete --ignore-not-found=$(ignore-not-found) -f deployment/sangagh/crds

.PHONY: deploy
deploy: manifests kustomize ## Deploy sanjagh to the K8s cluster specified in ~/.kube/config.
	helmsman -apply -f ./deployment/helmsman.yaml

.PHONY: undeploy
undeploy: ## Undeploy sanjagh from the K8s cluster specified in ~/.kube/config.
	helmsman -destroy -f ./deployment/helmsman.yaml

##@ Build Dependencies

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v3.8.7
CONTROLLER_TOOLS_VERSION ?= v0.11.1

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || { curl -Ss $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: operator-sdk
OPERATOR_SDK ?= $(LOCALBIN)/operator-sdk
operator-sdk: ## Download operator-sdk locally if necessary.
ifeq (,$(wildcard $(OPERATOR_SDK)))
ifeq (, $(shell which operator-sdk 2>/dev/null))
	@{ \
	set -e ;\
	mkdir -p $(dir $(OPERATOR_SDK)) ;\
	OS=$(shell go env GOOS) && ARCH=$(shell go env GOARCH) && \
	curl -sSLo $(OPERATOR_SDK) https://github.com/operator-framework/operator-sdk/releases/download/$(OPERATOR_SDK_VERSION)/operator-sdk_$${OS}_$${ARCH} ;\
	chmod +x $(OPERATOR_SDK) ;\
	}
else
OPERATOR_SDK = $(shell which operator-sdk)
endif
endif
