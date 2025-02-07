# AIIDA Helm

This is the repository, that holds a helm chart for the deployment of AIIDA using K3s.

## Tools

Overview of the tools used in this repository:

| Tool           | Description                            |
|----------------|----------------------------------------|
| kubectl        | K8s command line tool                  |
| K3s            | Lightweight K8s distribution           |
| Helm           | K8s package manager                    |
| Sealed Secrets | K8s secret management                  |
| K9s            | optional for visualizing k8s resources |
| yq             | YAML processor                         |

## Install AIIDA on Kubernetes Cluster
Set Namespace as temporary variable in local environment:
```bash
export NAMESPACE=mynamespace
```

Create initial values file:
```bash
touch chart/values-$NAMESPACE.yaml
```

The `setup-arch.bash` script is designed to set up all the necessary tools for Arch. Before running the script, you need to create a personal access token in GitHub with the `read:packages` scope. The username should be your GitHub username.

The script performs the following tasks:

1. Installs the required tools.
2. Creates sealed secrets for all usernames and passwords, including a sealed secret containing the personal access token. These sealed secrets are used for securely pulling the `aiida` image from the GitHub Container Registry and for other authentication purposes.

To customize the configuration, copy the `.values-template.yaml` file and edit the values as needed.

To execute the script, replace YOUR_USERNAME and YOUR_PERSONAL_ACCESS_TOKEN with your GitHub username and personal access token, then run the following command:
```bash
./setup-arch.bash --ghcr-username="YOUR_USERNAME" --ghcr-pat="YOUR_PERSONAL_ACCESS_TOKEN"
```

If you are planning to install multiple aiida instances on the same cluster, add this to your `values.yaml`:
```yaml
sealed-secrets:
  namespace: "hardening"
  fullnameOverride: "sealed-secrets-controller-hardening"
  rbac:
    create: true
    clusterRoleName: "hardening-secrets-unsealer"
```
