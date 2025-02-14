package helm

import (
	"context"
	"github.com/eddie-energy/aiida-helm/installer/internal/dto"
	helm "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"log"
	"strings"
	"time"
)

const timeout = 30 * time.Minute
const repositoryCache = "/tmp/.helmcache"
const repositoryConfig = "/tmp/.helmrepo"

func GetChartSpec(
	namespace string,
	repository repo.Entry,
	releaseName string,
	chartName string,
	valuesOptions values.Options,
	wait bool,
) helm.ChartSpec {
	return helm.ChartSpec{
		ReleaseName:     releaseName,
		ChartName:       getFullChartName(repository.Name, chartName),
		Namespace:       namespace,
		CreateNamespace: true,
		ValuesOptions:   valuesOptions,
		SkipCRDs:        true,
		Wait:            wait,
		Timeout:         timeout,
	}
}

func GetVersionInfosRepo(repository repo.Entry, namespace string, releaseNameSuffix *string) ([]*dto.VersionInfoDto, error) {
	helmClient, err := getHelmClient(namespace)
	if err != nil {
		return nil, err
	}

	if err := addAndUpdateChartRepos(helmClient, repository); err != nil {
		return nil, err
	}

	chartReleases, err := helmClient.ListReleasesByStateMask(action.ListAll)
	if err != nil {
		return nil, err
	}

	var versionInfos []*dto.VersionInfoDto
	for _, chartRelease := range chartReleases {
		if releaseNameSuffix == nil || strings.HasSuffix(chartRelease.Name, *releaseNameSuffix) {
			chartName := getFullChartName(repository.Name, chartRelease.Chart.Name())

			versionInfo := buildVersionInfoDto(chartRelease)
			versionInfo.LatestChart, err = getLatestChartMetadata(helmClient, chartName)
			if err != nil {
				return nil, err
			}

			versionInfos = append(versionInfos, versionInfo)
		}
	}

	return versionInfos, nil
}

func GetVersionInfo(repository repo.Entry, chartSpec helm.ChartSpec) (*dto.VersionInfoDto, error) {
	helmClient, err := getHelmClient(chartSpec.Namespace)
	if err != nil {
		return nil, err
	}

	if err := addAndUpdateChartRepos(helmClient, repository); err != nil {
		return nil, err
	}

	chartRelease, err := helmClient.GetRelease(chartSpec.ReleaseName)
	if err != nil {
		return nil, err
	}

	versionInfo := buildVersionInfoDto(chartRelease)
	versionInfo.LatestChart, err = getLatestChartMetadata(helmClient, chartSpec.ChartName)

	return versionInfo, err
}

func InstallOrUpgrade(repository repo.Entry, chartSpec helm.ChartSpec) (*dto.VersionInfoDto, error) {
	helmClient, err := getHelmClient(chartSpec.Namespace)
	if err != nil {
		return nil, err
	}

	if err := addAndUpdateChartRepos(helmClient, repository); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	chartRelease, err := helmClient.InstallOrUpgradeChart(ctx, &chartSpec, nil)
	if err != nil {
		return nil, err
	}

	return buildVersionInfoDto(chartRelease), nil
}

func Uninstall(chartSpec helm.ChartSpec) error {
	helmClient, err := getHelmClient(chartSpec.Namespace)
	if err != nil {
		return err
	}

	return helmClient.UninstallRelease(&chartSpec)
}

func getHelmClient(namespace string) (helm.Client, error) {
	helmOptions := &helm.Options{
		Namespace:        namespace,
		RepositoryCache:  repositoryCache,
		RepositoryConfig: repositoryConfig,
		Debug:            true,
		Linting:          true,
		DebugLog:         func(format string, v ...interface{}) { log.Printf(format+"\n", v...) },
	}

	return helm.New(helmOptions)
}

func addAndUpdateChartRepos(helmClient helm.Client, repository repo.Entry) error {
	if err := helmClient.AddOrUpdateChartRepo(repository); err != nil {
		return err
	}

	return helmClient.UpdateChartRepos()
}

func buildVersionInfoDto(release *release.Release) *dto.VersionInfoDto {
	return &dto.VersionInfoDto{
		ReleaseName:    release.Name,
		ReleaseInfo:    release.Info,
		InstalledChart: release.Chart.Metadata,
	}
}

func getLatestChartMetadata(helmClient helm.Client, chartName string) (*chart.Metadata, error) {
	latestChart, _, err := helmClient.GetChart(chartName, &action.ChartPathOptions{})
	if err != nil {
		return nil, err
	}

	return latestChart.Metadata, nil
}

func getFullChartName(repoName string, chartName string) string {
	return repoName + "/" + chartName
}
