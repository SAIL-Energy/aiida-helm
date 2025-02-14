package dto

import (
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

type VersionInfoDto struct {
	ReleaseName    string          `json:"release_name,omitempty"`
	ReleaseInfo    *release.Info   `json:"release_info,omitempty"`
	InstalledChart *chart.Metadata `json:"installed_chart,omitempty"`
	LatestChart    *chart.Metadata `json:"latest_chart,omitempty"`
}
