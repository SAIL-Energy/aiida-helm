# AIIDA Deployment

> [!WARNING]
> This Helm chart is not actively maintained and may not be up to date with the latest AIIDA features or deployment best practices.

This repository contains all the necessary resources to install/upgrade/delete AIIDA on a Kubernetes cluster.

## Repository Components

The repository includes the following components:

- **Helm chart for AIIDA** - Defines the Kubernetes resources for AIIDA.
- **Helm repository for AIIDA** - Hosts and manages Helm charts.

---

## Helm Chart

The **Helm chart** is located in the `chart/` directory and contains all required Kubernetes resources for deploying
AIIDA Core.

> [!NOTE]
> All commands should be executed from the root of this repository.

### Prerequisites

Before deploying the Helm chart, ensure you have the following prerequisites (K3s, kubectl, ...) installed:

```bash
bash <(curl -sfL https://eddie-energy.github.io/aiida-helm/setup.bash)
```

### Generate Secrets

Secrets are sealed using [Bitnami Sealed Secrets](https://github.com/bitnami-labs/sealed-secrets) to ensure secure storage and management.
To generate the necessary secrets for the Helm chart, run the following script.

Secrets can be adapted to your needs by modifying the file located at `chart/secrets-template.yaml`.

Before running the script, create a **GitHub Personal Access Token** with the `read:packages` scope.

To execute the script, replace `YOUR_USERNAME` and `YOUR_PERSONAL_ACCESS_TOKEN` with your credentials, then run:

```bash
bash <(curl -sfL https://eddie-energy.github.io/aiida-helm/setup-secrets.bash) --ghcr-username="YOUR_USERNAME" --ghcr-pat="YOUR_PERSONAL_ACCESS_TOKEN"
```

### Install local chart for testing

To test changes made to the Helm chart, follow these steps:

1. **Uninstall the existing release**

   Before reinstalling the chart manually, uninstall the chart and delete all persistent volumes:

   ```bash
   make uninstall
   make delete-all-pvcs
   ```

2. **Install the chart manually with Helm**

   Once cleanup is complete, you can install the Helm chart manually:
    
   ```bash
   make install
   ```
---

### Install AIIDA using Helm

To install AIIDA on your Kubernetes cluster using Helm, run the following command from the root of this repository:

```bash
helm install aiida \
  --namespace "$NAMESPACE" \
  --values ./chart/secrets.yaml \
  aiida/aiida
```

## Helm Repository

The **Helm repository** is located in the `repository/` directory and hosts AIIDA's Helm charts. The repository is:

- Hosted on **GitHub Pages**.
- Automatically updated when changes are pushed to the `main` branch.

### Adding a New Chart/Version to the Repository

To add a new version of the Helm chart, follow these steps:

1. **Package the chart** by running the following command in the root of this repository:

   ```bash
   make package CHART_PATH=./chart TRACK=stable
   ```
   
   - `CHART_PATH` defines the location of the chart
   - `TRACK` specifies the release track:
     - `stable` – For stable releases.
     - `beta` – For beta releases.

2. **Update the repository index** after packaging:

   ```bash
   make index TRACK=stable
   ```

   This updates the `index.yaml` file in the `repository/` directory.
   The `TRACK` variable determines whether the chart is added to the `stable` or `beta` track.