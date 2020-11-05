#!/usr/bin/env sh

set -e

# Load GCS Service account credentials into kube
kubectl create secret generic gcs-creds --from-file ~/.config/gcloud/service-account.json

# Create another nginx deployment with a GCS output
kubectl apply -f tester-gcs.yaml
