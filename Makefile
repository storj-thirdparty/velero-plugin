K8S_DIR := .k8s
K8S_CLUSTER_NAME := tardigrade
KUBECONFIG = $(realpath $(K8S_DIR)/config)
BUCKET_NAME ?= velero
STORJ_ACCESS ?= # Required
VELERO_BACKUP_LOCATION := default

export KUBECONFIG

DOCKER_IMAGE := storjlabs/velero-plugin
BRANCH_NAME ?= $(shell git rev-parse --abbrev-ref HEAD | sed "s!/!-!g")
DOCKER_TAG := ${BRANCH_NAME}
# TODO: add --exclude "v[0-9]*\.[0-9]*\.[0-9]*[!0-9]*" after releasing v1.0.0
# to avoid updating :latest to Beta and RC releases.
ifneq (,$(shell git describe --tags --exact-match --match "v[0-9]*\.[0-9]*\.[0-9]*"))
DOCKER_TAG_LATEST := true
endif

define HELP_MSG
Run a reproducible local development environment to develop the Velero \
plugin for Tardigrade. Usage: make [target]
endef

.PHONY: help
help: ## show this help message
	@echo $(HELP_MSG)
	@cat Makefile | awk -F ":.*##"  '/##/ { printf "    %-30s %s\n", $$1, $$2 }' | grep -v  grep

.PHONY: k8s-start
k8s-start: ## start and initialize a local K8s cluster
	@if [ ! -f $(K8S_DIR)/config ]; then \
		mkdir -p $(K8S_DIR); \
		kind create cluster --kubeconfig $(K8S_DIR)/config --name $(K8S_CLUSTER_NAME) --wait 1m; \
	fi

.PHONY: velero-install
velero-install: k8s-start ## install Velero and this plugin into the K8s cluster
	@velero install --no-secret \
		--no-default-backup-location \
		--use-volume-snapshots=false \
		--wait

.PHONY: velero-plugin-add
velero-plugin-add: ## add this plugin to Velero server
	@velero plugin add $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: velero-plugin-remove
velero-plugin-remove: ## delete this plugin from Velero server
	@velero plugin remove $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: velero-backup-location-create
velero-backup-location-create: .is-access-set ## create a backup location using Tardigrade plugin
	@velero backup-location create $(VELERO_BACKUP_LOCATION) \
		--bucket=$(BUCKET_NAME) \
		--config accessGrant=$(STORJ_ACCESS) \
		--provider=tardigrade

.PHONY: velero-backup-location-delete
velero-backup-location-delete: ## delete the backup location that uses Tardigrade plugin
		@kubectl -n velero delete backupstoragelocation.velero.io $(VELERO_BACKUP_LOCATION)

.PHONY: velero-uninstall
velero-uninstall: ## uninstall Velero from the K8s cluster
ifeq ($(call is_k8s_running),)
	@kubectl delete namespace/velero clusterrolebinding/velero
	@kubectl delete crds -l component=velero
endif

.PHONY: plugin-build
plugin-build: ## build docker image of this Velero plugin for development
	@docker image build --rm -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: plugin-image-push
plugin-image-push: ## push the plugin image to the local  K8s cluster
	@kind load docker-image $(DOCKER_IMAGE):$(DOCKER_TAG) --name=$(K8S_CLUSTER_NAME)

.PHONY: go-build
go-build: ## build the Go source
	@CGO_ENABLED=0 go build -o velero-plugin-for-tardigrade ./cmd

.PHONY: test
test: ## Run tests on source code
	cd testsuite; go test -race -v -cover -coverprofile=.coverprofile ./...
	@echo done


.PHONY: push
push: plugin-build ## pushes the Docker image to Docker Hub
	@docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
ifeq ($(DOCKER_TAG_LATEST), true)
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest
	docker push $(DOCKER_IMAGE):latest
endif

.PHONY: dev-env-init
dev-env-init: k8s-start velero-install ## start K8s cluster and install Velero

.PHONY: dev-env-start
dev-env-start: plugin-build plugin-image-push velero-plugin-add ## build and push image, and install the plugin

.PHONY: dev-env-refresh
dev-env-refresh: velero-plugin-remove dev-env-start ## build the plugin again and reinstall it

.PHONY: dev-env-destroy
dev-env-destroy: ## destroy the local development environment
	-@kind delete cluster --name $(K8S_CLUSTER_NAME)
	-@rm -rf $(K8S_DIR)

.PHONY: clean
clean: dev-env-destroy ## clean up the local environment on your local machine
	-@docker image rm -f $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: .is-access-set
.is-access-set:
	@$(if $(STORJ_ACCESS),,$(error STORJ_ACCESS environment variable is required))
