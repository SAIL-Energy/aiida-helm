package dto

import "helm.sh/helm/v3/pkg/chart"

type ChartMetaDataDto struct {
	Name        string `json:"name,omitempty"`
	Version     string `json:"version,omitempty"`
	Description string `json:"description,omitempty"`
	AppVersion  string `json:"appVersion,omitempty"`
	Deprecated  bool   `json:"deprecated,omitempty"`
}

func NewChartMetaDataDtoFromHelmChartMetadata(metadata *chart.Metadata) *ChartMetaDataDto {
	return &ChartMetaDataDto{
		Name:        metadata.Name,
		Version:     metadata.Version,
		Description: metadata.Description,
		AppVersion:  metadata.AppVersion,
		Deprecated:  metadata.Deprecated,
	}
}
