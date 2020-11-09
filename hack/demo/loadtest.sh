#!/usr/bin/env sh

# Create an ubuntu deployment
kubectl create deployment ubuntu --image ubuntu:latest  && \
sleep 2 && \
kubectl patch deployment ubuntu --patch '{"spec": { "template": { "spec": { "containers": [ { "name": "ubuntu", "command": [ "/bin/bash"], "args": [ "-c", "apt update && apt install -y curl && sleep 123123123123" ] } ] } } } }'

# Wait for container to come up
sleep 5

kubectl exec -it $(kubectl get po -l app=ubuntu -o jsonpath='{.items[0].metadata.name}') -- /bin/bash -xc "while true; do date; curl -F vote=b -s $(kubectl get svc -l component=voting-service -o jsonpath='{.items[0].spec.clusterIP}');  done"
