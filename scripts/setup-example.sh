#!/bin/bash

#check the registry


minikube start --driver=virtualbox
kubectl create deployment --image=registry:2 registry
kubectl expose deployment registry --port=5000 

PRIVATE_REG=`kubectl get service/registry -o jsonpath='{.spec.clusterIP}'`:5000

echo "Install velero plugin..."
velero install --provider gcp --plugins $PRIVATE_REG/velero-plugin:v0.0.1 --bucket $BUCKET --backup-location-config accessGrant=$ACCESS --no-secret

echo "Create snapshot location..."
velero snapshot-location create test-snap --provider tardigrade.io/volume-snapshotter --config accessGrant=$ACCESS_SNAP

echo "Create backup..."
velero backup create nginx-test --include-namespaces nginx-example --snapshot-volumes --volume-snapshot-locations test-snap

