# Postgres

Tawny's primary database is installed via the `install` script in `/scripts`. It leverages helm and bitnami charts.

```kubernetes helm
helm upgrade --install postgres oci://registry-1.docker.io/bitnamicharts/postgresql \
    --set nameOverride=tawny \
    --namespace tawny \
    --values ./zarf/k8s/postgres/values.yaml
```
The secret is autogenerated by helm and printed at the conclusion of the `install` script.
