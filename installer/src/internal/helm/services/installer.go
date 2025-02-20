package services

import (
	"github.com/eddie-energy/aiida-helm/installer/internal/dto"
	helmInstaller "github.com/eddie-energy/aiida-helm/installer/internal/helm"
	"github.com/google/uuid"
	helm "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/repo"
)

const namespace = "aiida-services"

type Installer struct {
	Repository repo.Entry
}

func NewInstaller(repository repo.Entry) *Installer {
	return &Installer{Repository: repository}
}

func (serviceInstaller *Installer) GetVersionInfosRepo(userId *string) ([]*dto.VersionInfoDto, error) {
	return helmInstaller.GetVersionInfosRepo(serviceInstaller.Repository, namespace, userId)
}

func (serviceInstaller *Installer) GetVersionInfo(chartName string, releaseName string) (*dto.VersionInfoDto, error) {
	return helmInstaller.GetVersionInfo(serviceInstaller.Repository, serviceInstaller.getChartSpec(chartName, releaseName))
}

func (serviceInstaller *Installer) InstallNew(userId string, chartName string) (*dto.VersionInfoDto, error) {
	releaseId, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	releaseName := chartName + "-" + releaseId.String() + "-" + userId
	return helmInstaller.InstallOrUpgrade(serviceInstaller.Repository, serviceInstaller.getChartSpec(chartName, releaseName))
}

func (serviceInstaller *Installer) InstallOrUpgrade(chartName string, releaseName string) (*dto.VersionInfoDto, error) {
	return helmInstaller.InstallOrUpgrade(serviceInstaller.Repository, serviceInstaller.getChartSpec(chartName, releaseName))
}

func (serviceInstaller *Installer) Uninstall(chartName string, releaseName string) error {
	return helmInstaller.Uninstall(serviceInstaller.getChartSpec(chartName, releaseName))
}

func (serviceInstaller *Installer) getChartSpec(chartName string, releaseName string) helm.ChartSpec {
	return helmInstaller.GetChartSpec(namespace, serviceInstaller.Repository, releaseName, chartName, values.Options{}, false)
}
