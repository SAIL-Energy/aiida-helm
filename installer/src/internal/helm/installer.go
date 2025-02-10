package helm

import (
	"context"
	helm "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	"log"
	"time"
)

const repositoryCache = "/tmp/.helmcache"
const repositoryConfig = "/tmp/.helmrepo"

func GetChartSpec(namespace string, repository repo.Entry, chartName string, valuesOptions values.Options, wait bool) helm.ChartSpec {
	return helm.ChartSpec{
		ReleaseName:     chartName,
		ChartName:       getFullChartName(repository.Name, chartName),
		Namespace:       namespace,
		CreateNamespace: true,
		ValuesOptions:   valuesOptions,
		SkipCRDs:        true,
		Wait:            wait,
		Timeout:         30 * time.Minute,
	}
}

func GetVersionInfosRepo(repository repo.Entry, namespace string) ([]*VersionInfo, error) {
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

	var versionInfos []*VersionInfo
	for _, chartRelease := range chartReleases {
		chartName := getFullChartName(repository.Name, chartRelease.Chart.Name())

		versionInfo := buildVersionInfo(chartRelease)
		versionInfo.LatestChart, err = getLatestChartMetadata(helmClient, chartName)
		if err != nil {
			return nil, err
		}

		versionInfos = append(versionInfos, versionInfo)
	}

	return versionInfos, nil
}

func GetVersionInfo(repository repo.Entry, chartSpec helm.ChartSpec) (*VersionInfo, error) {
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

	versionInfo := buildVersionInfo(chartRelease)
	versionInfo.LatestChart, err = getLatestChartMetadata(helmClient, chartSpec.ChartName)

	return versionInfo, err
}

func Install(repository repo.Entry, chartSpec helm.ChartSpec) (*VersionInfo, error) {
	helmClient, err := getHelmClient(chartSpec.Namespace)
	if err != nil {
		return nil, err
	}

	if err := addAndUpdateChartRepos(helmClient, repository); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	chartRelease, err := helmClient.InstallChart(ctx, &chartSpec, nil)
	if err != nil {
		return nil, err
	}

	return buildVersionInfo(chartRelease), nil
}

func Upgrade(repository repo.Entry, chartSpec helm.ChartSpec) (*VersionInfo, error) {
	helmClient, err := getHelmClient(chartSpec.Namespace)
	if err != nil {
		return nil, err
	}

	if err := addAndUpdateChartRepos(helmClient, repository); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	chartRelease, err := helmClient.UpgradeChart(ctx, &chartSpec, nil)
	if err != nil {
		return nil, err
	}

	return buildVersionInfo(chartRelease), nil
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

func buildVersionInfo(release *release.Release) *VersionInfo {
	return &VersionInfo{
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
