package dto

import (
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/time"
)

type ReleaseInfoDto struct {
	FirstDeployed time.Time      `json:"firstDeployed,omitempty"`
	LastDeployed  time.Time      `json:"lastDeployed,omitempty"`
	Deleted       time.Time      `json:"deleted"`
	Description   string         `json:"description,omitempty"`
	Status        release.Status `json:"status,omitempty"`
}

func NewReleaseInfoDtoFromHelmReleaseInfo(releaseInfo *release.Info) *ReleaseInfoDto {
	return &ReleaseInfoDto{
		FirstDeployed: releaseInfo.FirstDeployed,
		LastDeployed:  releaseInfo.LastDeployed,
		Deleted:       releaseInfo.LastDeployed,
		Description:   releaseInfo.Description,
		Status:        releaseInfo.Status,
	}
}
