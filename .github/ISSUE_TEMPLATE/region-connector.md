---
name: Region Connector
about: Define a new region-connector
title: '<country> <name of datahub/permission admin/metered data admin> region-connector'
labels: ["needs-approval", "requirement"]
assignees: ''
projects: ["eddie-energy/1"]
---

## Short Description

_Describe the region-connector, where to find the documentation, credentials, contact person, etc_

## Region-Connector Essentials

The following things have to be implemented to consider the region-connector complete.
The list is intended to be converted to issues,
where each issue is one or multiple points.

- [ ] Create a permission request at the permission administrators side
- [ ] Implement custom element for region-connector
- [ ] Implement permission market documents
- [ ] Request validated historical data and emit it to raw data stream
- [ ] Map validated historical data to validated historical data market documents
- [ ] Allow data needs for future data, and request the data once available
- [ ] Request accounting point data and emit it to raw data stream
- [ ] Map accounting point data to accounting point market document
- [ ] Ensure that data needs are enforced by region-connector, such as only requesting the correct data.
For example, not requesting gas metered data for data need that specifies electricity.
- [ ] React to revocation of permission request
- [ ] Allow termination of permission requests
  - [ ] Remove credentials if possible
  - [ ] Terminate on the permission administrators side if possible
- [ ] Time out stale permission requests
- [ ] Implement health indicators for external APIs and services

## Blockers

- [ ] _please list any blockers here that need to be finished so that this issue can be tackled_

