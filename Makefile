
# Image URL to use all building/pushing image targets
REGISTRY ?= casbin
IMG ?= casdoor
IMG_TAG ?=$(shell git --no-pager log -1 --format="%ad" --date=format:"%Y%m%d")-$(shell git describe --tags --always --dirty --abbrev=6)
NAMESPACE ?= casdoor
APP ?= casdoor
HOST ?= test.com

MYSQL_ROOT_PASSWORD ?= 123456
MYSQL_DATA ?= /usr/local/docker/mysql
CONF_DIR ?= ./conf
CREATE_DATABASE ?= true

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: docker-build docker-push deploy

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

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: ut
ut: ## UT test
	go test -v -cover -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

##@ Build

.PHONY: backend
backend: fmt vet ## Build backend binary.
	go build -o bin/manager main.go

.PHONY: backend-vendor
backend-vendor: vendor fmt vet ## Build backend binary with vendor.
	go build -mod=vendor -o bin/manager main.go

.PHONY: frontend
frontend: ## Build backend binary.
	cd web/ && yarn && yarn run build && cd -

.PHONY: vendor
vendor: ## Update vendor.
	go mod vendor

.PHONY: run
run: fmt vet ## Run backend in local 
	go run ./main.go

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build -t ${REGISTRY}/${IMG}:${IMG_TAG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${REGISTRY}/${IMG}:${IMG_TAG}

.PHONY: docker-build-standard
docker-build-standard: ## Build docker standard image with the docker compose.
	docker-compose build casdoor

.PHONY: docker-run-standard
docker-run-standard: ## Run with docker compose.
	CASDOOR_IMG=${REGISTRY}/${IMG} CASDOOR_TAG=${IMG_TAG} \
	CONF_DIR=${CONF_DIR} \
	MYSQL_DATA=${MYSQL_DATA} MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD} \
	CREATE_DATABASE=${CREATE_DATABASE} \
	docker-compose up

lint-install: ## Install golangci-lint
	@# The following installs a specific version of golangci-lint, which is appropriate for a CI server to avoid different results from build to build
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.40.1

lint: ## Run golangci-lint
	@echo "---lint---"
	golangci-lint run --modules-download-mode=vendor ./...

##@ Deployment

.PHONY: deploy
deploy:  ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	helm upgrade --install ${APP} manifests/casdoor --create-namespace --set ingress.enabled=true \
	--set "ingress.hosts[0].host=${HOST},ingress.hosts[0].paths[0].path=/,ingress.hosts[0].paths[0].pathType=ImplementationSpecific" \
	--set image.tag=${IMG_TAG} --set image.repository=${REGISTRY} --set image.name=${IMG} --version ${IMG_TAG} -n ${NAMESPACE}

.PHONY: dry-run
dry-run: ## Dry run for helm install
	helm upgrade --install ${APP} manifests/casdoor --set ingress.enabled=true \
	--set "ingress.hosts[0].host=${HOST},ingress.hosts[0].paths[0].path=/,ingress.hosts[0].paths[0].pathType=ImplementationSpecific" \
	--set image.tag=${IMG_TAG} --set image.repository=${REGISTRY} --set image.name=${IMG} --version ${IMG_TAG} -n ${NAMESPACE} --dry-run

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	helm delete ${APP} -n ${NAMESPACE}
