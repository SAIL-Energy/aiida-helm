# AIIDA Deployment

This repository contains all the necessary resources to install/upgrade/delete AIIDA and its services on a Kubernetes cluster.

## Repository Components

The repository includes the following components:

- **Helm chart for AIIDA** – Defines the Kubernetes resources for AIIDA.
- **Helm repository for AIIDA and its services** – Hosts and manages Helm charts.
- **AIIDA Installer** – The tool to install AIIDA and its associated services.

---

## Helm Chart

The **Helm chart** is located in the `chart/` directory and contains all required Kubernetes resources for deploying
AIIDA Core.

> **Note:**  
> The chart is installed via the AIIDA Installer and is **not** meant to be installed manually.

---

## Helm Repository

The **Helm repository** is located in the `repository/` directory and hosts AIIDA's Helm charts. The repository is:

- Hosted on **GitHub Pages**.
- Automatically updated when changes are pushed to the `main` branch.

### Adding a New Chart/Version to the Repository

To add a new version of the Helm chart, follow these steps:

1. **Package the chart** by running the following command in the root of this repository:

   ```bash
   make package CHART_PATH=./chart TRACK=beta
   ```
   
   - `CHART_PATH` defines the location of the chart (e.g., `installer` or `core`).
   - `TRACK` specifies the release track:
     - `stable` – For stable releases.
     - `beta` – For beta releases.

2. **Update the repository index** after packaging:

   ```bash
   make index TRACK=beta
   ```

   This updates the `index.yaml` file in the `repository/` directory.
   The `TRACK` variable determines whether the chart is added to the `stable` or `beta` track.

---

## AIIDA Installer

The **AIIDA Installer** simplifies the installation of AIIDA and its services on a Kubernetes cluster. Services are installed via a marketplace.

### Installer Directory Structure

- `chart/` – Contains the Helm chart for the installer.
- `src/` – Contains the Golang source code for the installer.

### Installation Steps

To install AIIDA on a **K3s** cluster, follow these steps:

1. Install **K3s** on the cluster.
2. Install **kubectl** on the local machine.
3. Set up a **GitHub Container Registry (GHCR)** token.
4. Install **Helm**.
5. Install the **AIIDA Installer** with Helm on the cluster.

The `setup.bash` script is provided to automate the setup process on Linux.

#### Running the Setup Script

Before running the script, create a **GitHub Personal Access Token** with the `read:packages` scope.

To execute the script, replace `YOUR_USERNAME` and `YOUR_PERSONAL_ACCESS_TOKEN` with your credentials, then run:

```bash
./setup.bash --ghcr-username="YOUR_USERNAME" --ghcr-pat="YOUR_PERSONAL_ACCESS_TOKEN"
```

### Customizing the Installation

Like any Helm chart, the AIIDA installation can be customized by providing a `values.yaml` file.

Customization options include:

1. **Installer Settings** – Modify parameters for the installer itself, such as the repository track.
2. **AIIDA Core Settings** – Adjust AIIDA core configurations in the `aiida` section (e.g., change the port).
3. **Secrets** – Add secrets to the `secrets` section (e.g., add a password for the NanoMQ service). Otherwise, the secrets will be generated automatically.

#### Example `values.yaml` Configuration:

```yaml
core:
  repositoryTrack: "stable"

aiida:
  core:
    port: 8888

secrets:
  nanomq:
    encryptedPwd: "admin: supersafe"
```

#### Applying Customizations

To use a custom configuration, pass the `values.yaml` file using the `--values` flag when installing AIIDA:

```bash
helm install aiida-installer \
  --namespace "$NAMESPACE" \
  --values values.yaml \
  aiida/aiida-installer
```

This ensures that AIIDA is deployed with the specified custom settings.