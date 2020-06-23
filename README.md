# velero plugin for storj

Velero is a tool to backup Kubernetes clusters. Velero does two things: 1) performs backups of the workloads running in a kubernetes cluster and 2) takes snapshots of persistent volumes.

Velero is able to integrate with a variety of storage systems plugins. There are two types of plugins available: object store or volume snapshotter (or both). 

Here we implement a Velero Object Store Plugin that is backed by Storj object storage.

## install velero with storj object store plugin

#### prerequisites

- Complete Velero Prerequisites and install the CLI: [docs](https://velero.io/docs/master/basic-install/)
- Create a storj tardigrade account: [docs](https://tardigrade.io/signup/)
- Created a project in the storj account: [docs](https://documentation.tardigrade.io/getting-started/uploading-your-first-object/create-a-project)
- Create an api key for the project: [docs](https://documentation.tardigrade.io/getting-started/uploading-your-first-object/create-an-api-key)
- Setup a storj Uplink CLI and create an access grant for the project: [docs](https://documentation.tardigrade.io/getting-started/uploading-your-first-object/set-up-uplink-cli)
- Create a storj bucket where Velero will store the backups: [docs](https://documentation.tardigrade.io/getting-started/uploading-your-first-object/create-a-bucket)  

#### install velero with storj plugin

```
$ velero install --provider gcp \
    --plugins storjthirdparty/velero-plugin:v0.1.0 \
    --bucket $BUCKET \
    --backup-location-config accessGrant=$ACCESS \
    --no-secret
```

#### backup/restore

Perform a backup:

```
$ velero backup create $BACKUP_NAME
```

Perform a restore:

```
$ velero restore create $RESTORE_NAME --from-backup $BACKUP_NAME
```

### steps to publish a new velero plugin for storj

```
$ docker build -t storjthirdparty/velero-plugin:<version> .

$ docker push storjthirdparty/velero-plugin:<version>
```
