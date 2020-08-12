# Velero plugin for Storj

Velero is a tool to backup Kubernetes clusters. Velero does two things:

1. performs backups of the workloads running in a kubernetes cluster and,
1. takes snapshots of persistent volumes.

Velero is able to integrate with a variety of storage systems plugins. There are two types of plugins available: object store or volume snapshotter (or both).

Here we implement a Velero Object Store Plugin that is backed by Storj object storage.

## Install Velero with Storj object store plugin

### Prerequisites

- Complete Velero Prerequisites and install the CLI: [docs](https://velero.io/docs/master/basic-install/)
- Create a storj tardigrade account: [docs](https://tardigrade.io/signup/)
- Created a project in the storj account: [docs](https://documentation.tardigrade.io/getting-started/uploading-your-first-object/create-a-project)
- Create an api key for the project: [docs](https://documentation.tardigrade.io/getting-started/uploading-your-first-object/create-an-api-key)
- Setup a storj Uplink CLI and create an access grant for the project: [docs](https://documentation.tardigrade.io/getting-started/uploading-your-first-object/set-up-uplink-cli)
- Create a storj bucket where Velero will store the backups: [docs](https://documentation.tardigrade.io/getting-started/uploading-your-first-object/create-a-bucket)

### Install velero with Storj plugin

```
$ velero install --provider gcp \
    --plugins storjthirdparty/velero-plugin:v0.1.0 \
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

NOTE Kind has some [known issues](https://kind.sigs.k8s.io/docs/user/known-issues/) and some of them are related with the OS and architecture.


### Development

For development you'll also need [Go](https://golang.org/) installed.

The `Makefile` contains a list of targets with the prefix `dev-env-`. The purpose of them are for easing the development cycle. For example, there the `dev-env-refresh` target that uninstall, build and install again the plugin, which is useful when you make changes in the plugin source code.


### Considerations

The local Kubernetes cluster created with Kind by the `Makefile` creates a Kubernetes configuraiton file in `.k8s` directory.

For connecting to this cluster with kubectl or Velero CLI you need to export the `KUBECONFIG` environment variable pointing to such file (e.g. `export KUBECONFIG=$(pwd)/.k8s/config`).

## steps to publish a new velero plugin for storj

```
$ docker build -t storjthirdparty/velero-plugin:<version> .

$ docker push storjthirdparty/velero-plugin:<version>
```
