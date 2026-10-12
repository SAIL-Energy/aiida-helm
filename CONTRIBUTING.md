# Contributing to the AIIDA Helm chart

FlexiData GmbH developed this chart. Use the [chart repository](https://github.com/eddie-energy/aiida-helm) for chart issues and
pull requests. Report issues in the AIIDA application to
the [EDDIE project](https://github.com/eddie-energy/eddie).

Describe the configuration that triggers the issue, the chart and application
versions, the expected result, and the actual behaviour. Include a small
sanitised values file where possible. Remove credentials and Secret data from
rendered manifests before sharing them.

For a change, explain its effect on existing installations. Update the packaged
README and examples when their behaviour changes.
Keep the common values file readable and document advanced settings where users
need them. Consider the effect of resource names and storage changes on populated installations.

Follow the checks in [RELEASE.md](RELEASE.md) before proposing a release.
Published chart versions are immutable. Increment the chart version for every
published change, and use a new major version for incompatible configuration
or resource changes. A chart version change does not automatically upgrade AIIDA.

Contributions are accepted under the [Apache License, Version 2.0](LICENSE).
Retain applicable attribution notices. Report security issues privately using
[SECURITY.md](SECURITY.md).

Developer: FlexiData GmbH, [www.flexidata.at](https://www.flexidata.at),
[office@flexidata.at](mailto:office@flexidata.at).
