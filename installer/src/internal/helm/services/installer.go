package services

import (
	"fmt"
	"github.com/eddie-energy/aiida-helm/installer/internal/dto"
	helmInstaller "github.com/eddie-energy/aiida-helm/installer/internal/helm"
	"github.com/eddie-energy/aiida-helm/installer/internal/helm/aiida"
	helm "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/repo"
	"strings"
)

const namespace = "aiida-services"
const maxReleasesPerChart = 1000

type Installer struct {
	Repository repo.Entry
}

func NewInstaller(repository repo.Entry) *Installer {
	return &Installer{Repository: repository}
}

func (serviceInstaller *Installer) GetVersionInfosRepo(userId *string, chartName *string) ([]*dto.VersionInfoDto, error) {
	return helmInstaller.GetVersionInfosRepo(serviceInstaller.Repository, namespace, chartName, userId)
}

func (serviceInstaller *Installer) GetVersionInfo(chartName string, releaseName string) (*dto.VersionInfoDto, error) {
	chartSpec := serviceInstaller.getChartSpec(chartName, releaseName, nil)
	return helmInstaller.GetVersionInfo(serviceInstaller.Repository, chartSpec)
}

func (serviceInstaller *Installer) InstallNew(userId string, chartName string, setupDto *dto.SetupDto) (*dto.VersionInfoDto, error) {
	if err := verifyChartName(chartName); err != nil {
		return nil, err
	}

	versionInfos, err := serviceInstaller.GetVersionInfosRepo(&userId, &chartName)
	if err != nil {
		return nil, err
	}

	releaseName := generateReleaseName(userId, chartName, versionInfos)
	chartSpec := serviceInstaller.getChartSpec(chartName, releaseName, setupDto)

	return helmInstaller.InstallOrUpgrade(serviceInstaller.Repository, chartSpec)
}

func (serviceInstaller *Installer) InstallOrUpgrade(chartName string, releaseName string) (*dto.VersionInfoDto, error) {
	if err := verifyChartName(chartName); err != nil {
		return nil, err
	}

	chartSpec := serviceInstaller.getChartSpec(chartName, releaseName, nil)
	return helmInstaller.InstallOrUpgrade(serviceInstaller.Repository, chartSpec)
}

func (serviceInstaller *Installer) Uninstall(chartName string, releaseName string) error {
	chartSpec := serviceInstaller.getChartSpec(chartName, releaseName, nil)
	return helmInstaller.Uninstall(chartSpec)
}

func (serviceInstaller *Installer) getChartSpec(chartName string, releaseName string, setupDto *dto.SetupDto) helm.ChartSpec {
	valuesOptions := values.Options{}
	if setupDto != nil {
		valuesOptions.Values = setupDto.CustomValues
	}

	return helmInstaller.GetChartSpec(
		namespace,
		serviceInstaller.Repository,
		releaseName, chartName,
		valuesOptions,
		false,
	)
}

func verifyChartName(chartName string) error {
	if chartName == aiida.ChartName || chartName == helmInstaller.InstallerChartName {
		return fmt.Errorf("the chart name '%s' is reserved and cannot be used as a service", chartName)
	}
	return nil
}

func generateReleaseName(userId string, chartName string, versions []*dto.VersionInfoDto) string {
	releaseNum := len(versions) + 1
	var releaseId string

	for ; releaseNum < maxReleasesPerChart; releaseNum++ {
		releaseId = fmt.Sprintf("%d-%s", releaseNum, userId)
		if !containsReleaseId(versions, releaseId) {
			break
		}
	}

	maxChartNameLength := helmInstaller.MaxReleaseNameLength - len(releaseId) - 1
	if len(chartName) > maxChartNameLength {
		chartName = chartName[:maxChartNameLength]
	}

	return fmt.Sprintf("%s-%s", chartName, releaseId)
}

func containsReleaseId(versions []*dto.VersionInfoDto, releaseId string) bool {
	for _, version := range versions {
		if strings.Contains(version.ReleaseName, releaseId) {
			return true
		}
	}
	return false
}
