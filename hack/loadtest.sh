#!/usr/bin/env sh

set -e

# Create an ubuntu deployment
kubectl create deployment ubuntu --image ubuntu:latest  && \
sleep 2 && \
kubectl patch deployment ubuntu --patch '{"spec": { "template": { "spec": { "containers": [ { "name": "ubuntu", "command": [ "/bin/bash"], "args": [ "-c", "sleep 123123123123" ] } ] } } } }'

# Run curl at 1req/s
kubectl exec -it $(kubectl get po -l app=ubuntu -o jsonpath='{.items[0].metadata.name}') -- /bin/bash -xc "apt update && apt install -y curl && while true; do date; curl -s $(kubectl get po -l app=funsies -o jsonpath='{.items[0].status.podIP }'); sleep 1; done"
