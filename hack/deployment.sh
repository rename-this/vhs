#!/usr/bin/env sh

set -e

# kind load docker-image sidecarinjector:latest
# kind load docker-image us.gcr.io/carbon-relay-dev/vhs:latest

kubectl apply -f namespace.yaml

./create_certs.sh --service vhs-webhook --namespace redsky-system --secret webhook-tls

 cacert=$(kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}')

 sed -i 's/caBundle:.*/caBundle: '${cacert}'/' mutatingwebhook.yaml

 kustomize build . | kubectl apply -f -

 kubectl label namespace default vhs="" --overwrite=true

 kubectl apply -f tester.yaml
