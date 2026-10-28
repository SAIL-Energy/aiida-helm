# AIIDA Helm chart

This chart deploys [AIIDA](https://architecture.eddie.energy/aiida/), the EDDIE project's Administrative Interface for In-House Data Access, on Kubernetes. It includes AIIDA core, TimescaleDB, and a choice of Mosquitto or HiveMQ for MQTT. Keycloak is an external prerequisite.

FlexiData GmbH developed this chart. The chart is a separate project from the [upstream EDDIE framework](https://github.com/eddie-energy/eddie).

**Current chart version: <!-- CHART_VERSION_START -->`1.1.0`<!-- CHART_VERSION_END -->.**

[The chart guide](chart/README.md) is the authoritative installation and configuration documentation. It includes an installation using existing Secrets, public access, storage, workload controls, and verification commands. Start there and create a small private values file.

Each component runs as one instance. The default AIIDA image is `v1.52.2`, pinned by digest. The internal example uses port forwarding for the UI; public HTTP requires an ingress controller, DNS, and TLS. External MQTT is disabled by default. Review the configured images and verify the intended application flows before using the deployment for production.

| Documentation | Purpose |
| --- | --- |
| [Chart guide](chart/README.md) | Install and operate the chart |
| [Keycloak guide](chart/docs/KEYCLOAK.md) | Configure the external identity provider |
| [Public MQTT](chart/README.md#mqtt-and-external-devices) | Select a broker and configure public access and certificate renewal |
| [Database operation](chart/README.md#database-operation) | Configure an external database and understand backup scope |
| [Release guide](RELEASE.md) | Prepare and publish immutable chart packages |
| [Contributing](CONTRIBUTING.md) | Report chart issues and propose changes |
| [Security policy](SECURITY.md) | Report vulnerabilities privately |

Source: [eddie-energy/aiida-helm](https://github.com/eddie-energy/aiida-helm). Follow the source installation in the chart guide. For a published OCI package, use the registry reference announced for that release. The package includes the chart README, examples, and Keycloak guide.

The [AIIDA operation manual](https://architecture.eddie.energy/aiida/) explains application behaviour and supported data sources. This chart supplies Kubernetes configuration. Identity administration, device access, and the complete backup and recovery procedure remain the operator's responsibility.

The chart was developed by FlexiData GmbH and is distributed under the [Apache License 2.0](LICENSE). The deployed applications and their dependencies retain their own licences. See [NOTICE](NOTICE) for attribution.

[FlexiData GmbH](https://www.flexidata.at)

Hammerschmidtgasse 5, 1190 Vienna, Austria

[office@flexidata.at](mailto:office@flexidata.at)
