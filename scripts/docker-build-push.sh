#!/bin/bash
# This script pushes the plugin image to the kubernetes docker registry
# Requirements:
# - docker registry has to be set up in kubernetes (using minikube)
# - a version tag must be defined in git

PLUGIN_VERSION=`git describe --abbrev=0 --tags`
DOCKERFILE_DIR=/hosthome/fadila/dev/velero/velero-plugin
PRIVATE_REG=`kubectl get service/registry -o jsonpath='{.spec.clusterIP}'`:5000

minikube ssh "docker build $DOCKERFILE_DIR -t $PRIVATE_REG/velero-plugin:$PLUGIN_VERSION"
minikube ssh "docker push $PRIVATE_REG/velero-plugin:$PLUGIN_VERSION"
