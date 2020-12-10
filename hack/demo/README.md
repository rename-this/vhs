# VHS Demo


# Create GCS Service Account secret
We'll need to load a GCS service account into a kubernetes as a secret.
This file should be named `service-account.json`.

```bash
# Load GCS Service account credentials into kube
kubectl create secret generic gcs-creds --from-file service-account.json
```

# Deploy application

```bash
kustomize build . | kubectl apply -f -
```

# Run load tester

```bash
./loadtest.sh
```
