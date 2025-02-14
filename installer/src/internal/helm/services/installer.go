package services

import (
	"github.com/eddie-energy/aiida-helm/installer/internal/dto"
	helmInstaller "github.com/eddie-energy/aiida-helm/installer/internal/helm"
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

func (serviceInstaller *Installer) GetVersionInfo(userId string, chartName string) (*dto.VersionInfoDto, error) {
	chartSpec := serviceInstaller.getChartSpec(userId, chartName)

	return helmInstaller.GetVersionInfo(serviceInstaller.Repository, chartSpec)
}

func (serviceInstaller *Installer) InstallOrUpgrade(userId string, chartName string) (*dto.VersionInfoDto, error) {
	return helmInstaller.InstallOrUpgrade(serviceInstaller.Repository, serviceInstaller.getChartSpec(userId, chartName))
}

func (serviceInstaller *Installer) Uninstall(userId string, chartName string) error {
	return helmInstaller.Uninstall(serviceInstaller.getChartSpec(userId, chartName))
}

func (serviceInstaller *Installer) getChartSpec(userId string, chartName string) helm.ChartSpec {
	releaseName := chartName + "-" + userId
	return helmInstaller.GetChartSpec(namespace, serviceInstaller.Repository, releaseName, chartName, values.Options{}, false)
}
