# velero plugin for storj tardigrade

Velero is a tool to backup Kubernetes clusters. Velero does two things: 1) performs backups of the workloads running in a kubernetes cluster and 2) takes snapshots of persistent volumes.

Velero is able to integrate with a variety of storage systems plugins. There are two types of plugins available: object store or volume snapshotter (or both). 

Here we implement a Velero Object Store Plugin that is backed by Storj Tardigrade object storage.

### Install velero with storj tardigrade object store plugin

```
$ velero install --provider gcp \
    --plugins jessgreb01/velero-plugin-for-storj-tardigrade:v0.0.2 \
    --bucket $BUCKET \
    --backup-location-config accessGrant=$ACCESS \
    --no-secret
```

Perform a backup:

```
$ velero backup create $BACKUP_NAME
```

Perform a restore:

```
$ velero restore create $RESTORE_NAME --from-backup $BACKUP_NAME
```

### Steps to publish a new velero plugin for storj tardigrade

```
$ docker build -t jessgreb01/velero-plugin-for-storj-tardigrade:v0.0.2 .

$ docker push jessgreb01/velero-plugin-for-storj-tardigrade:v0.0.2
```
