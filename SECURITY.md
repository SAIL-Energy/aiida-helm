# Security reporting

Report suspected vulnerabilities privately to FlexiData GmbH at
[office@flexidata.at](mailto:office@flexidata.at). Include the affected chart and
application versions, the configuration involved, and steps that demonstrate
the issue. Do not include production credentials, personal data, or Secret contents.

Use a private report rather than a public issue for an undisclosed vulnerability.
Identify whether the issue appears to belong to this chart, AIIDA, or another
component so that the report can reach the relevant project.

The chart records its application and dependency versions in `chart/Chart.yaml`
and the values files. Check those versions when assessing whether a vulnerability
affects an installation. This project does not promise a maintenance period or
a response deadline. Any published fixes use new chart versions. Existing
published versions are never replaced.

The chart's security settings depend on the images and cluster policies in use.
Review the packaged [security and workload guidance](chart/README.md#security-and-workload-controls)
when choosing a configuration. Keep the Kubernetes cluster, AIIDA, database,
and broker images within the versions supported by their maintainers.

Scan the exact configured image references before deployment and after changing
them. Resolve findings according to the deployment's security
policy, and verify the application's authentication and authorization behaviour.

Developer: FlexiData GmbH, [www.flexidata.at](https://www.flexidata.at).
