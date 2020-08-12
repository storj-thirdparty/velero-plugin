K8S_DIR := .k8s
K8S_CLUSTER_NAME := storj
KUBECONFIG = $(realpath $(K8S_DIR)/config)
BUCKET_NAME ?= velero
STORJ_ACCESS ?= # Required
DOCKER_IMAGE := storjthirdparty/velero-plugin
DOCKER_TAG ?= dev
VELERO_BACKUP_LOCATION := default

export KUBECONFIG

define HELP_MSG
Run a reproducible local development environment to develop the Storj Velero \
plugin. Usage: make [target]
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
velero-backup-location-create: .is-access-set ## create a backup location using Storj plugin
	@velero backup-location create $(VELERO_BACKUP_LOCATION) \
		--bucket=$(BUCKET_NAME) \
		--config accessGrant=$(STORJ_ACCESS) \
		--provider=tardigrade

.PHONY: velero-backup-location-delete
velero-backup-location-delete: ## delete the backup location that uses Storj plugin
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
	@CGO_ENABLED=0 go build -o velero-plugin-storj .

.PHONY: dev-env-init
dev-env-init: k8s-start velero-install ## start K8s cluster and install Velero

.PHONY: dev-env-start
dev-env-start: plugin-build plugin-image-push velero-plugin-add ## build and push image, and install the plugin

.PHONY: dev-env-refresh
dev-env-refresh: velero-plugin-remove dev-env-start ## build the plugin again and reinstall it

.PHONY: dev-env-destroy
dev-env-destroy: ## destroy the local development environment
	@kind delete cluster --name $(K8S_CLUSTER_NAME) || true
	@rm -rf $(K8S_DIR)

.PHONY: clean
clean: dev-env-destroy ## clean up the local environment on your local machine
	@docker image rm -f $(DOCKER_IMAGE):$(DOCKER_TAG) || true

.PHONY: .is-access-set
.is-access-set:
	@$(if $(STORJ_ACCESS),,$(error STORJ_ACCESS environment variable is required))
