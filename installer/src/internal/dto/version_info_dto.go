package dto

import "helm.sh/helm/v3/pkg/release"

type VersionInfoDto struct {
	ReleaseName    string            `json:"releaseName,omitempty"`
	ReleaseInfo    *ReleaseInfoDto   `json:"releaseInfo,omitempty"`
	InstalledChart *ChartMetaDataDto `json:"installedChart,omitempty"`
	LatestChart    *ChartMetaDataDto `json:"latestChart,omitempty"`
}

func NewVersionInfoDtoFromHelmRelease(release *release.Release) *VersionInfoDto {
	return &VersionInfoDto{
		ReleaseName:    release.Name,
		ReleaseInfo:    NewReleaseInfoDtoFromHelmReleaseInfo(release.Info),
		InstalledChart: NewChartMetaDataDtoFromHelmChartMetadata(release.Chart.Metadata),
	}
}
