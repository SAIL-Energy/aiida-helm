# Releasing the chart

Publish reviewed chart packages manually. Each chart version identifies immutable contents. Publishing does not upgrade running installations.

## Prepare the release

The source repository is [eddie-energy/aiida-helm](https://github.com/eddie-energy/aiida-helm). The source installation is documented in [the chart guide](chart/README.md). OCI distribution is optional. Choose the registry and namespace before publishing a package, then announce the complete chart reference after verifying retrieval.

`chart/Chart.yaml: version` is the package version. Use a patch for compatible fixes, a minor for compatible additions, and a major for incompatible configuration or deployment changes. `appVersion` describes AIIDA and stays aligned with the configured core image.

Keep the chart version, root README version marker, packaged version references, and versioned commands consistent. Use plain SemVer such as `1.0.0`. Never replace a published package or move its release tag.

Confirm application and dependency compatibility, licences, notices, and image availability. Scan the exact configured image references, including HiveMQ and both backup containers when those features are offered. Review the findings under the deployment's security policy.

## Validate and package

Use a clean checkout of the reviewed release commit. From the repository root:

```sh
git status --short
helm lint ./chart --strict -f chart/examples/values.internal.yaml
helm lint ./chart --strict -f chart/examples/values.public.yaml
helm lint ./chart --strict -f chart/examples/values.internal.yaml --set mqtt.broker=hivemq
helm lint ./chart --strict -f chart/examples/values.public.yaml -f chart/examples/values.mqtt-public.yaml
helm lint ./chart --strict -f chart/examples/values.internal.yaml -f chart/examples/values.database-external.yaml
helm lint ./chart --strict -f chart/examples/values.internal.yaml -f chart/examples/values.workload.yaml
helm template aiida ./chart -n aiida -f chart/examples/values.internal.yaml > rendered.local.yaml
kubeconform -strict -summary -kubernetes-version 1.27.0 rendered.local.yaml
mkdir -p dist
helm package ./chart --destination dist
tar -tzf dist/aiida-1.0.0.tgz
```

The chart has no Helm dependencies to build. Repeat rendering and Kubernetes schema validation for the configurations being distributed, including the selected broker, database TLS, public access, and backup settings.

Review the rendered resources and archive contents. Include the README, licence, notices, values schema, examples, and Keycloak guide. Keep private values, rendered Secrets, keys, and registry credentials outside the package. Git ignore patterns do not replace `.helmignore`.

Helm linting and rendering assess chart output. Verify image retrieval, application startup, Keycloak login, MQTT authentication and topic authorization, certificate renewal, and complete recovery separately before accepting the release. A passing manifest check does not resolve image dependency findings.

Record the reviewed source commit with an annotated or signed tag:

```sh
git show HEAD:chart/Chart.yaml
git tag -a chart-v1.0.0 HEAD -m "AIIDA chart 1.0.0"
git push origin refs/tags/chart-v1.0.0
```

Confirm that the tag, package filename, and chart metadata contain the selected version. The tag records the source commit and does not publish the package.

## Publish to a chosen registry

Replace the example host and namespace below with the registry selected for distribution. These are placeholders, not an existing chart package location:

```sh
REGISTRY_HOST="registry.example.org"
REGISTRY_NAMESPACE="oci://registry.example.org/organisation/charts"
CHART_REF="$REGISTRY_NAMESPACE/aiida"
helm registry login "$REGISTRY_HOST"
```

Use credentials with permission to publish to that namespace. Before pushing, check that the version is absent. This Bash block permits publication only when the registry explicitly reports a missing manifest. An existing version, authentication failure, connectivity failure, or unrecognized response stops publication:

```sh
ERROR_FILE="$(mktemp)"
if helm show chart "$CHART_REF" --version 1.0.0 >/dev/null 2>"$ERROR_FILE"; then
  printf 'Version 1.0.0 already exists. Select a new chart version.\n' >&2
  rm "$ERROR_FILE"
  false
elif grep -Eqi 'manifest unknown|MANIFEST_UNKNOWN' "$ERROR_FILE"; then
  rm "$ERROR_FILE"
  helm push dist/aiida-1.0.0.tgz "$REGISTRY_NAMESPACE"
else
  cat "$ERROR_FILE" >&2
  rm "$ERROR_FILE"
  false
fi
```

If the registry uses a different missing manifest response, verify absence through its own interface before publishing. Other errors do not establish that a version is available to use. Coordinate publication and configure registry version immutability when available.

The destination of `helm push` omits the chart name and tag because Helm reads both from the package. See [Helm's OCI registry guidance](https://helm.sh/docs/topics/registries/). Configure public package access before announcing an anonymous download location.

## Verify retrieval

After publishing, use the actual `CHART_REF` with a clean Helm registry configuration to confirm anonymous access:

```sh
REGISTRY_CONFIG="$(mktemp)"
printf '{"auths":{}}\n' > "$REGISTRY_CONFIG"
helm show chart "$CHART_REF" --version 1.0.0 --registry-config "$REGISTRY_CONFIG"
helm show readme "$CHART_REF" --version 1.0.0 --registry-config "$REGISTRY_CONFIG"
rm "$REGISTRY_CONFIG"
```

Download and inspect the published package. Record the registry digest and verification outcome. Confirm anonymous retrieval of every image advertised as public, including the HiveMQ reference when offering that broker.

If publication fails, first check whether the package now exists. Correct access or connectivity before retrying. Source changes require a new chart version. Never overwrite an existing release to correct its contents.
