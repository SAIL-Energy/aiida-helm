# AIIDA Helm chart

This chart deploys AIIDA core, an optional bundled TimescaleDB database, and one MQTT broker. Mosquitto is the default broker. HiveMQ is an optional integration using the EDDIE authentication plugin image. Keycloak is configured separately.

FlexiData GmbH developed this chart as a separate project from the [EDDIE framework](https://github.com/eddie-energy/eddie). Read the [AIIDA operation manual](https://architecture.eddie.energy/aiida/) for product concepts and supported data sources.

## Versions and deployment scope

| Component | Chart default or requirement |
| --- | --- |
| Chart | `1.0.0` |
| AIIDA core | `v1.52.2`, pinned by digest |
| Kubernetes | `1.27` or later |
| Helm | `3.18` or later |
| Bundled database | TimescaleDB `2.30.2` on PostgreSQL `17` |
| Identity provider | External Keycloak realm and public OpenID Connect client |
| MQTT broker | Mosquitto or HiveMQ |

Each component supports one instance. The chart does not configure high availability. AIIDA persists its keystore and MQTT client state. The bundled database and selected broker also use persistent storage.

Ingress is enabled in the default values and requires an explicit hostname. The internal example disables it and uses port forwarding. External MQTT is disabled by default. Changing application or image versions requires compatibility and image availability checks.

## Install with existing Secrets

The following installation provides an internal deployment with a local UI URL. It uses existing Secrets so that credentials are not stored in Helm values. Commands use Bash and require `helm`, `kubectl`, and `openssl`. PowerShell users can put commands on one line or use PowerShell's backtick continuation. A secret management system can create equivalent Secrets with the keys listed below.

Use the same release name and namespace throughout. The examples use `aiida`. Run source installation commands from the repository root, which contains `chart/`.

Obtain the chart source and enter its repository root before continuing:

```sh
git clone --branch open-source-handover https://github.com/eddie-energy/aiida-helm.git
cd aiida-helm
```

For a reproducible installation, use the selected source commit or an available release tag.

### 1. Check the cluster

You need permission to create namespaced workloads and persistent volumes through an available StorageClass:

```sh
kubectl config current-context
kubectl cluster-info
kubectl get storageclass
kubectl create namespace aiida --dry-run=client -o yaml | kubectl apply -f -
```

A default StorageClass is sufficient. Set `core.persistence.storageClass`, `timescaledb.storageClass`, and `mqtt.persistence.storageClass` explicitly when the cluster has no suitable default. The chart does not install an ingress controller, cert-manager, or Keycloak.

### 2. Configure Keycloak

Create a Keycloak realm and public OpenID Connect client with matching `realm` and `clientId` values. For this internal installation, allow redirect URI `http://localhost:8080/*` and web origin `http://localhost:8080`. The provider must be reachable by both AIIDA and the browser.

The [Keycloak guide](docs/KEYCLOAK.md) describes provider configuration, HTTPS, PKCE, and verification. Use your actual provider URL in the values file.

### 3. Create the credential Secret

Generate separate passwords for the database, broker database role, MQTT administrator, and AIIDA keystore:

```sh
DATABASE_PASSWORD="$(openssl rand -hex 24)"
BROKER_DATABASE_PASSWORD="$(openssl rand -hex 24)"
MQTT_PASSWORD="$(openssl rand -hex 24)"
KEYSTORE_PASSWORD="$(openssl rand -hex 32)"

kubectl -n aiida create secret generic aiida-runtime-credentials \
  --from-literal=username=aiida_user \
  --from-literal=password="$DATABASE_PASSWORD" \
  --from-literal=databasePassword="$BROKER_DATABASE_PASSWORD" \
  --from-literal=mqttPassword="$MQTT_PASSWORD" \
  --from-literal=keystorePassword="$KEYSTORE_PASSWORD"
```

Store the generated passwords in your password manager. The keystore and its original password must remain available together for as long as their data does. These commands create a new Secret; they do not rotate passwords in an existing database or re-encrypt a keystore.

All referenced Secrets must exist in the release namespace before installation. Keep private values and rendered manifests outside version control.

### 4. Create a small private values file

Copy [the internal installation example](examples/values.internal.yaml) and replace its Keycloak URL:

```sh
cp chart/examples/values.internal.yaml values.local.yaml
```

The complete example contains:

```yaml
ingress:
  enabled: false
core:
  publicUrl: http://localhost:8080
  keystore:
    existingSecret: aiida-runtime-credentials
    existingSecretKey: keystorePassword
timescaledb:
  existingSecret: aiida-runtime-credentials
mqtt:
  broker: mosquitto
  existingSecret: aiida-runtime-credentials
keycloak:
  externalHost: https://keycloak.example.com
  realm: AIIDA
  clientId: AIIDA
```

Change the realm and client ID if yours differ. Internal UI access needs no public DNS, ingress controller, or cert-manager. External devices need the public MQTT configuration described below.

### 5. Validate and install from source

From the repository root, run:

```sh
helm lint ./chart --strict -f values.local.yaml
helm template aiida ./chart -n aiida -f values.local.yaml > rendered.local.yaml
helm upgrade --install aiida ./chart --namespace aiida \
  -f values.local.yaml --wait --timeout 10m
```

Inspect resource names, images, Secret references, PVCs, and public endpoints before installing. The chart has no Helm dependencies to build. Rendered output can contain credentials when inline fallback values are used.

### Optional: install a packaged chart

The package includes its examples and Keycloak guide. Install a local package with:

```sh
helm upgrade --install aiida ./aiida-1.0.0.tgz --namespace aiida \
  -f values.local.yaml --wait --timeout 10m
```

For an announced OCI release, set `CHART_REF` to its complete registry reference and select the version explicitly:

```sh
helm upgrade --install aiida "$CHART_REF" --version 1.0.0 \
  --namespace aiida -f values.local.yaml --wait --timeout 10m
```

Confirm that the package is available and can be retrieved with the intended registry access.

### 6. Verify the deployment

Helm reporting `STATUS: deployed` does not establish that the application's identity integration and data sources work. Inspect workloads and storage:

```sh
helm status aiida -n aiida
kubectl -n aiida get pods,pvc,service,ingress
kubectl -n aiida rollout status deployment/aiida-core --timeout=5m
kubectl -n aiida rollout status statefulset/aiida-timescaledb --timeout=5m
kubectl -n aiida rollout status statefulset/aiida-mosquitto --timeout=5m
kubectl -n aiida port-forward service/aiida-core 8080:8080
```

Keep port forwarding running. In another terminal, check readiness:

```sh
curl --fail --show-error http://localhost:8080/actuator/health/readiness
```

Open `http://localhost:8080` and complete a Keycloak login. Exercise the intended data source and authorized MQTT traffic before accepting the deployment. Skip the TimescaleDB workload command for an external database. The broker StatefulSet is named `aiida-hivemq` when HiveMQ is selected.

## Public access

### HTTP and Keycloak

Public HTTP requires DNS, an ingress controller, and a TLS certificate. The default ingress class is `traefik`. Chart managed certificates require cert-manager and an existing ClusterIssuer such as `letsencrypt-prod`.

For a new public installation, copy [the public HTTP example](examples/values.public.yaml) to `values.local.yaml`. For an existing internal configuration, replace its application URL and enable ingress:

```yaml
core:
  publicUrl: ""
ingress:
  enabled: true
  className: traefik
  host: aiida.example.com
  tls:
    enabled: true
    clusterIssuer: letsencrypt-prod
```

Point the AIIDA DNS record at the ingress controller. Update the Keycloak redirect URI to `https://aiida.example.com/*` and web origin to `https://aiida.example.com`. Check the relevant prerequisites:

```sh
kubectl get ingressclass
kubectl get crd certificates.cert-manager.io
kubectl get clusterissuer letsencrypt-prod
```

To use an existing certificate instead, set `ingress.tls.clusterIssuer: ""` and `ingress.tls.existingSecret` to a `kubernetes.io/tls` Secret containing `tls.crt` and `tls.key`. cert-manager is then unnecessary for this ingress.

### MQTT and external devices

Enable external MQTT only when devices outside the cluster need it. Merge [the public MQTT example](examples/values.mqtt-public.yaml) over your installation values:

```sh
helm upgrade --install aiida ./chart --namespace aiida \
  -f values.local.yaml -f chart/examples/values.mqtt-public.yaml --wait --timeout 10m
```

Replace its hostname first. Create the referenced TLS Secret and point DNS at the MQTT LoadBalancer. Public TLS uses port 8883. Public plaintext on port 1883 requires a separate explicit opt-in through `mqtt.externalAccess.allowPlaintext`.

Configure the endpoint before creating data sources for external devices. Changing the chart does not rewrite endpoints already stored by AIIDA. If the MQTT hostname or TLS Secret name is empty, the chart reuses the web hostname or certificate Secret respectively. The certificate must cover the hostname used by clients.

Mosquitto authenticates clients and authorizes topics through AIIDA's database. To select HiveMQ instead, set `mqtt.broker: hivemq`. Its configured public image is `ghcr.io/eddie-energy/hivemq-auth-plugin:sha-d085114`; no pull Secret is required. The chart supplies the AIIDA-compatible database queries and credentials.

Mosquitto reloads changed mounted TLS certificates automatically. HiveMQ reads its certificate at startup. After renewing its TLS Secret, restart the broker and verify the served certificate:

```sh
kubectl -n aiida rollout restart statefulset/aiida-hivemq
kubectl -n aiida rollout status statefulset/aiida-hivemq
```

## Values and credentials

[values.yaml](values.yaml) documents the configuration interface. [values.schema.json](values.schema.json) rejects unsupported types and combinations. Copy only the keys you want to change into your private values file. Nested maps merge; lists replace the complete default list. Image `digest` takes precedence over `tag`; clear the digest when deliberately selecting a different image tag.

| Example | Purpose |
| --- | --- |
| [Internal installation](examples/values.internal.yaml) | Existing credentials and local UI access |
| [Public HTTP](examples/values.public.yaml) | Ingress with existing credential Secrets |
| [Public MQTT](examples/values.mqtt-public.yaml) | TLS endpoint for external devices; overlay |
| [External database](examples/values.database-external.yaml) | Verified TimescaleDB connection; overlay |
| [Workload controls](examples/values.workload.yaml) | Kubernetes resource request and limit shapes; overlay |

The internal and public HTTP examples are complete base configurations. Merge overlays after the selected base file with a second `-f` argument. Replace example hostnames and create any extra Secrets before installation.

| Setting | Behaviour |
| --- | --- |
| Application URL | Derived from ingress when enabled; otherwise `core.publicUrl` is required |
| MQTT public hostname | `mqtt.externalAccess.host`, then `ingress.host` |
| MQTT TLS Secret | `mqtt.externalAccess.tls.existingSecret`, then the effective web TLS Secret |
| Database connection | Bundled TimescaleDB, or `timescaledb.host` and `port` when disabled |
| Image selection | A configured digest takes precedence over the image tag |

### Existing Secrets

The chart omits its generated credential Secret when the corresponding `existingSecret` is set. Secret references may share one Secret, as in the installation above, or use separate Secrets.

| Values path | Required Secret keys |
| --- | --- |
| `core.keystore.existingSecret` | Key selected by `existingSecretKey`, normally `password`; the example selects `keystorePassword` |
| `timescaledb.existingSecret` | Keys selected by `existingSecretKeys`, normally `username` and `password` |
| `mqtt.existingSecret` | Keys selected by `existingSecretKeys`, normally `mqttPassword` and `databasePassword` |
| `timescaledb.tls.existingSecret` | Public CA certificate selected by `caKey`, normally `ca.crt` |
| `timescaledb.backup.s3.existingSecret` | Keys selected by `existingSecretKeys`, normally `accessKey` and `secretKey` |
| Web or MQTT TLS Secret | `tls.crt` and `tls.key` |

Inline credentials are stored in Helm release records. Prefer existing Secrets. Mosquitto database passwords must not contain line breaks; the generated hexadecimal passwords above meet this requirement.

Changing a credential Secret does not restart workloads automatically or alter passwords already stored in PostgreSQL. Coordinate database role password changes, Secret updates, and affected workload restarts. A keystore password change requires stopping AIIDA, backing up the matching keystore and database, and re-encrypting the existing keystore before using the new password. Merely updating the Secret makes the existing keystore unreadable.

## Database operation

AIIDA requires database `aiida` and broker account `emqx`. These names are fixed application requirements; the account name does not select an EMQX broker. Mosquitto and HiveMQ authenticate against AIIDA's tables.

To use an external TimescaleDB, set `timescaledb.enabled: false`, configure its hostname and credentials, and prefer `sslMode: verify-full` with a trusted CA Secret. Use [the external database example](examples/values.database-external.yaml) as an overlay. The database must support PostgreSQL 17, the required TimescaleDB extension, schema initialization, and broker grants. Provision `emqx` with the configured broker password, database CONNECT, schema USAGE, and SELECT privileges on `public.data_source_mqtt_user` and `public.data_source_mqtt_acl`. The bundled database uses internal plaintext connections with `sslMode: disable`.

Changing a bootstrap Secret does not rotate passwords inside an initialized database.

### Database backups

The optional scheduled S3 backup contains a compressed custom-format PostgreSQL `.dump` archive. Configure it in your private values file:

```yaml
timescaledb:
  backup:
    enabled: true
    s3:
      region: eu-central-1
      bucket: your-aiida-backup-bucket
      existingSecret: aiida-backup-credentials
```

Create the referenced Secret with `accessKey` and `secretKey`. An empty S3 endpoint uses AWS S3; a custom endpoint must use HTTPS. The default schedule is daily at 02:00 UTC. The database dump must succeed before upload starts.

These backups cover the database only. Complete recovery also needs the matching AIIDA keystore, its original password, MQTT client files, and any required broker and Keycloak state. Capture a consistent recovery point and verify restoration in an isolated environment. The chart does not automate a complete recovery.

## Security and workload controls

| Component | Default persistence | Default memory request and limit |
| --- | --- | --- |
| AIIDA core | `1Gi` for keystore and MQTT client state | `512Mi`, `2Gi` |
| TimescaleDB | `10Gi` | `512Mi`, `2Gi` |
| MQTT broker | `1Gi` | `256Mi`, `1Gi` |

These defaults are starting points. Adjust CPU, memory, storage, and broker limits for the expected workload. Resources use standard Kubernetes `requests` and `limits`; see [the workload example](examples/values.workload.yaml). Storage expansion depends on the provider and controller.

AIIDA runs as user and group 10001, with a read-only root filesystem, dropped capabilities, and a separate writable temporary volume. Core uses a `Recreate` update strategy so two instances do not write its persistent files at once. Other components use image-specific identities and permissions. HiveMQ needs writable image paths and an AMD64 node; do not copy another component's security context without checking image compatibility.

Core uses health probes; broker TCP probes check listener availability. Readiness does not verify Keycloak login, database privileges, topic authorization, or complete recovery. Registry credentials are unnecessary for the configured public images, including the supplied HiveMQ image.

AIIDA, Mosquitto, database, and backup images are pinned by digest. HiveMQ uses a commit tag; set its image digest when pinning an exact registry artifact. Verify image availability, scan the configured references, and review findings according to the deployment's security policy.

The core claim and chart-created keystore Secret are retained on uninstall by default through `core.persistence.retain`. Database and broker claims also remain. Retention is not a backup. Preserve the core volume and its original keystore password together with a consistent database backup.

## Troubleshooting and uninstalling

| Symptom | First check |
| --- | --- |
| Pending pod or PVC | StorageClass, provisioner events, quota, and access mode |
| Missing credential Secret | Namespace, Secret name, configured keys, and controller readiness |
| `ImagePullBackOff` | Repository availability, exact image reference, and anonymous registry access |
| Keycloak redirect or login failure | Application URL, issuer, realm, client ID, redirect URIs, and web origins |
| Database or broker authentication failure | Stored role passwords, grants, database TLS, and Secret rotation ordering |
| Missing or invalid TLS certificate | DNS, issuer readiness, Secret contents, hostname coverage, and broker reload |
| Keystore cannot be opened | Original password, restored file, ownership, and matching database recovery point |
| Backup failure | Dump and upload logs, Job deadline, staging storage, and S3 permissions |

```sh
kubectl -n aiida get events --sort-by=.lastTimestamp
kubectl -n aiida describe pod POD_NAME
kubectl -n aiida logs deployment/aiida-core
```

Uninstall with `helm uninstall aiida --namespace aiida`. Inspect retained state before deliberately deleting it. To reuse retained core storage and its keystore password, set `core.persistence.existingClaim` and `core.keystore.existingSecret`. Preserve the original password and matching database state. Helm rollback does not restore persistent data or reverse application database changes.

## Licence and project contact

The chart was developed by FlexiData GmbH and is distributed under the [Apache License 2.0](LICENSE). Applications and dependencies retain their own licences. See [NOTICE](NOTICE) for attribution.

Source and chart issues: [eddie-energy/aiida-helm](https://github.com/eddie-energy/aiida-helm).

Developer: [FlexiData GmbH](https://www.flexidata.at), Hammerschmidtgasse 5, 1190 Vienna, Austria.

Contact: [office@flexidata.at](mailto:office@flexidata.at). Report vulnerabilities privately to this address. Developer attribution and contact information do not establish a maintenance or support commitment.
