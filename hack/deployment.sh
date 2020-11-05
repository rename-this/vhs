#!/usr/bin/env sh

set -e

# kind load docker-image sidecarinjector:latest
# kind load docker-image us.gcr.io/carbon-relay-dev/vhs:latest

# Create namespace for sidecar injector ( using redsky-system )
kubectl apply -f namespace.yaml

# Create TLS certificates for webhook controller ( sidecar injector )
./create_certs.sh --service vhs-webhook --namespace redsky-system --secret webhook-tls

# Get CA certificate from cluster
cacert=$(kubectl config view --raw --minify --flatten -o jsonpath='{.clusters[].cluster.certificate-authority-data}')

# Update mutating webhook to use cluster CA bundle
sed -i 's/caBundle:.*/caBundle: '${cacert}'/' mutatingwebhook.yaml

# Create necessary resources for sidecar injector
kustomize build . | kubectl apply -f -

# Label the default namespace with `vhs` so we can target it with the sidecar injector
kubectl label namespace default vhs="" --overwrite=true

# Create a simple nginx deployment
kubectl apply -f tester.yaml
