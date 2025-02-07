package service

import (
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

func (serviceInstaller *Installer) GetVersionInfo() ([]*helmInstaller.VersionInfo, error) {
	return helmInstaller.GetVersionInfo(serviceInstaller.Repository, namespace)
}

func (serviceInstaller *Installer) GetSingleVersionInfo(chartName string) (*helmInstaller.VersionInfo, error) {
	chartSpec := serviceInstaller.getChartSpec(chartName)

	return helmInstaller.GetSingleVersionInfo(serviceInstaller.Repository, chartSpec)
}

func (serviceInstaller *Installer) Install(chartName string) error {
	return helmInstaller.Install(serviceInstaller.Repository, serviceInstaller.getChartSpec(chartName))
}

func (serviceInstaller *Installer) Upgrade(chartName string) error {
	return helmInstaller.Upgrade(serviceInstaller.Repository, serviceInstaller.getChartSpec(chartName))
}

func (serviceInstaller *Installer) Uninstall(chartName string) error {
	return helmInstaller.Uninstall(serviceInstaller.getChartSpec(chartName))
}

func (serviceInstaller *Installer) getChartSpec(chartName string) helm.ChartSpec {
	return helmInstaller.GetChartSpec(namespace, serviceInstaller.Repository, chartName, values.Options{})
}
