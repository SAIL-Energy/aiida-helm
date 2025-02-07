package helm

import (
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
)

type VersionInfo struct {
	ReleaseName    string          `json:"releaseName"`
	ReleaseInfo    *release.Info   `json:"releaseInfo"`
	InstalledChart *chart.Metadata `json:"installedChart"`
	AvailableChart *chart.Metadata `json:"availableChart"`
}
