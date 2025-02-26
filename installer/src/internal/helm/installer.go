package helm

import (
	"context"
	"github.com/eddie-energy/aiida-helm/installer/internal/dto"
	helm "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/repo"
	"log"
	"strings"
	"time"
)

const InstallerChartName = "aiida-installer"
const MaxReleaseNameLength = 53

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
		SkipCRDs:        false,
		Wait:            wait,
		Timeout:         timeout,
	}
}

func GetVersionInfosRepo(repository repo.Entry, namespace string, chartName *string, releaseNameSuffix *string) ([]*dto.VersionInfoDto, error) {
	log.Printf("Getting version infos for repository %s in namespace %s", repository.Name, namespace)

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

	versionInfos := make([]*dto.VersionInfoDto, 0)
	for _, chartRelease := range chartReleases {
		if chartName == nil || chartRelease.Chart.Name() == *chartName {
			if releaseNameSuffix == nil || strings.HasSuffix(chartRelease.Name, *releaseNameSuffix) {
				chartName := getFullChartName(repository.Name, chartRelease.Chart.Name())

				versionInfo := dto.NewVersionInfoDtoFromHelmRelease(chartRelease)
				versionInfo.LatestChart, err = getLatestChartMetadata(helmClient, chartName)
				if err != nil {
					return nil, err
				}

				versionInfos = append(versionInfos, versionInfo)
			}
		}
	}

	return versionInfos, nil
}

func GetVersionInfo(repository repo.Entry, chartSpec helm.ChartSpec) (*dto.VersionInfoDto, error) {
	log.Printf("Getting version info for chart %s in namespace %s", chartSpec.ChartName, chartSpec.Namespace)

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

	versionInfo := dto.NewVersionInfoDtoFromHelmRelease(chartRelease)
	versionInfo.LatestChart, err = getLatestChartMetadata(helmClient, chartSpec.ChartName)

	return versionInfo, err
}

func InstallOrUpgrade(repository repo.Entry, chartSpec helm.ChartSpec) (*dto.VersionInfoDto, error) {
	log.Printf("Installing or upgrading chart %s in namespace %s", chartSpec.ChartName, chartSpec.Namespace)

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

	return dto.NewVersionInfoDtoFromHelmRelease(chartRelease), nil
}

func Uninstall(chartSpec helm.ChartSpec) error {
	log.Printf("Uninstalling chart %s in namespace %s", chartSpec.ChartName, chartSpec.Namespace)

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

func getLatestChartMetadata(helmClient helm.Client, chartName string) (*dto.ChartMetaDataDto, error) {
	latestChart, _, err := helmClient.GetChart(chartName, &action.ChartPathOptions{})
	if err != nil {
		return nil, err
	}

	return dto.NewChartMetaDataDtoFromHelmChartMetadata(latestChart.Metadata), nil
}

func getFullChartName(repoName string, chartName string) string {
	return repoName + "/" + chartName
}
