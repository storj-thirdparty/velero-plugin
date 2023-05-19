# Velero plugin for Storj DCS

__IMPORTANT NOTICE__ we have archived this project because with our multi-tenant hosted S3 gateways
you can use [Velero AWS plugin](https://github.com/vmware-tanzu/velero-plugin-for-aws) to store the
data in Storj decentralized storage network.

Example configuration of Velero AWS plugin with Storj DCS.

```
velero install \
    --provider aws \
    --plugins velero/velero-plugin-for-aws:v1.6.0 \
    --bucket <YOUR_STORJ_BUCKET_NAME> \
    --backup-location-config region=global,s3Url=gateway.storjshare.io \
    --secret-file ./credentials-storj \
    --use-volume-snapshots=false \
    --use-restic
```


## Description

Velero is a tool to backup Kubernetes clusters. Velero does two things:

1. Performs backups of the workloads running in a kubernetes cluster and,
1. Takes snapshots of persistent volumes.

Velero is able to integrate with a variety of storage systems plugins. There are two types of plugins available: object store or volume snapshotter (or both).

Here we implement a Velero Object Store plugin that is backed by Storj DCS object storage.

## Installation

### Prerequisites

- Complete Velero prerequisites and install the CLI: [docs](https://velero.io/docs/main/basic-install/)
- Create a Storj DCS account: [docs](https://docs.storj.io/dcs/getting-started/quickstart-uplink-cli/uploading-your-first-object/prerequisites)
- Create an Access Grant: [docs](https://docs.storj.io/dcs/getting-started/quickstart-uplink-cli/uploading-your-first-object/create-first-access-grant)
- Set up Uplink CLI with the previously created Access Grant: [doc](https://docs.storj.io/dcs/getting-started/quickstart-uplink-cli/uploading-your-first-object/set-up-uplink-cli)
- Create a bucket where Velero will store the backups: [docs](https://docs.storj.io/dcs/getting-started/quickstart-uplink-cli/uploading-your-first-object/create-a-bucket)

### Install the Velero plugin for Tardigrade (a.k.a. Storj DCS)

```
$ velero install --provider tardigrade \
    --plugins storjlabs/velero-plugin \
    --bucket $BUCKET \
    --backup-location-config accessGrant=$ACCESS \
    --no-secret
```

### Backup & restore

Perform a backup:

```
$ velero backup create $BACKUP_NAME
```

Perform a restore:

```
$ velero restore create $RESTORE_NAME --from-backup $BACKUP_NAME
```

## Local environment

This repository contains a `Makefile` that expose several targets for creating, running and performing several Velero operations for easing the development of this plugin but also for being able to try it out.

Use `make help` to see the list of targets and a brief description of what they offer.


### Required tools

The local environment requires the following tools:

* [Kind](https://kind.sigs.k8s.io/): Local Kubernetes cluster using Docker containers. It requires to have [Docker](https://www.docker.com/products/docker-desktop) installed.
* [Kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/): The Kubernetes command-line tool.
* [Velero CLI](https://velero.io/docs/v1.4/basic-install/): The Velero command-line tool.

Be aware that the `Makefile` uses some Unix like tools that you may need to install depending of your Linux distribution and OSX and you will have to install for sure in Windows. While this local environment has been working fine in some Linux distributions and probably works in almost all of them, we believe that also works in OSX but we don't know if it will in Windows.

NOTE: Kind has some [known issues](https://kind.sigs.k8s.io/docs/user/known-issues/) and some of them are related with the OS and architecture.


### Development

For development you'll also need [Go](https://golang.org/) installed.

The `Makefile` contains a list of targets with the prefix `dev-env-`. The purpose of them are for easing the development cycle. For example, there the `dev-env-refresh` target that uninstall, build and install again the plugin, which is useful when you make changes in the plugin source code.


### Considerations

The local Kubernetes cluster created with Kind by the `Makefile` creates a Kubernetes configuraiton file in `.k8s` directory.

For connecting to this cluster with kubectl or Velero CLI you need to export the `KUBECONFIG` environment variable pointing to such file (e.g. `export KUBECONFIG=$(pwd)/.k8s/config`).

### How to publish a new version

Follow the [MAINTAINERS](MAINTAINERS.md) guide.
