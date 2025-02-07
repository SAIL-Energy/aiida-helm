package helm

import (
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

type VersionInfo struct {
	ReleaseName    string          `json:"release_name"`
	ReleaseInfo    *release.Info   `json:"release_info"`
	InstalledChart *chart.Metadata `json:"installed_chart"`
	LatestChart    *chart.Metadata `json:"latest_chart"`
}
