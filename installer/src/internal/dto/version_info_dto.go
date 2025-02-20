package dto

import (
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

type VersionInfoDto struct {
	ReleaseName    string          `json:"releaseName,omitempty"`
	ReleaseInfo    *release.Info   `json:"releaseInfo,omitempty"`
	InstalledChart *chart.Metadata `json:"installedChart,omitempty"`
	LatestChart    *chart.Metadata `json:"latestChart,omitempty"`
}
