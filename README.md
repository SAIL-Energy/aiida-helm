# AIIDA Helm

This is the repository, that holds a helm chart for the deployment of AIIDA using Kubernetes.

## TODOs

- [ ] Readiness Probes

## Tools

Overview of the tools used in this repository:

- K3s (lightweight k8s distribution)
- Helm (k8s package manager)
- Sealed Secrets (k8s secret management)
- K9s (optional for visualizing k8s resources)
- yq (YAML processor)

`setup.bash` is a script that installs all necessary tools for Arch.

## Install AIIDA on Kubernetes Cluster
Set Namespace as temporary variable in local environment:
```bash
export NAMESPACE=mynamespace
```

Create Namespace:
```bash
kubectl create namespace $NAMESPACE
```

Create initial values file:
```bash
touch aiida-chart/values-$NAMESPACE.yaml
```

Create git pull secret:
Retrieve github access token and set values: CR_USERNAME and CR_PAT
Recommendations regarding hist...
```bash
export CR_USERNAME=yourgithubusername
export CR_PAT=your generated pat
make create-ghcr-secret
```

add secret to values:
```yaml
ghcrImagePullSecrets:
  name: "ghcr-secret"
```

If you are planning to install multiple eddie instances on the same cluster, add this to your values-yaml:
```yaml
sealed-secrets:
  namespace: "hardening"
  fullnameOverride: "sealed-secrets-controller-hardening"
  rbac:
    create: true
    clusterRoleName: "hardening-secrets-unsealer"
```

Install
```bash
make install
```

### TimescaleDB
Create and encrypt TimescaleDB values.

Go to values-yournamespace.yaml and create values for:
```yaml
timescaledb:
  replicaCount: 1
  encryptedUsername: "your encrypted username in cleartext"
  encryptedPassword: "your encrypted password in cleartext"
```
Encrypt the values:
```bash
make seal-secret VALUE="timescaledb.encryptedUsername" SECRET_NAME="timescaledb-secret"
make seal-secret VALUE="timescaledb.encryptedPassword" SECRET_NAME="timescaledb-secret"
```

Upgrade Check if it works properly
```bash
make upgrade
```

Configure Ingress:
```bash
ingress:
  basePath: "univie"
  host: "eddie-worker.projekte.fh-hagenberg.at"
```

Create Cluster Issuer:
```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    email: developers@eddie.energy
    server: https://acme-v02.api.letsencrypt.org/directory
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
      - http01:
          ingress:
            class: traefik
```

Configure TLS:
```bash
ingress:
  tls: enabled
  clusterIssuer: "letsencrypt-prod"
```
