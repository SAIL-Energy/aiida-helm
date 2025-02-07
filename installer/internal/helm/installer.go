package helm

import (
	"context"
	helm "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/repo"
	"log"
	"time"
)

const repositoryCache = "/tmp/.helmcache"
const repositoryConfig = "/tmp/.helmrepo"

func GetChartSpec(namespace string, repository repo.Entry, chartName string, valuesOptions values.Options) helm.ChartSpec {
	return helm.ChartSpec{
		ReleaseName:     chartName,
		ChartName:       repository.Name + "/" + chartName,
		Namespace:       namespace,
		CreateNamespace: true,
		ValuesOptions:   valuesOptions,
		SkipCRDs:        true,
		Wait:            true,
	}
}

func GetVersionInfo(repository repo.Entry, namespace string) ([]*VersionInfo, error) {
	helmClient, err := getHelmClient(namespace)
	if err != nil {
		return nil, err
	}

	if err := addAndUpdateChartRepos(helmClient, repository); err != nil {
		return nil, err
	}

	releases, err := helmClient.ListReleasesByStateMask(action.ListAll)
	if err != nil {
		return nil, err
	}

	var versionInfos []*VersionInfo
	for _, release := range releases {
		versionInfo := &VersionInfo{
			ReleaseName:    release.Name,
			ReleaseInfo:    release.Info,
			InstalledChart: release.Chart.Metadata,
		}
		versionInfos = append(versionInfos, versionInfo)
	}

	return versionInfos, nil
}

func GetSingleVersionInfo(repository repo.Entry, chartSpec helm.ChartSpec) (*VersionInfo, error) {
	helmClient, err := getHelmClient(chartSpec.Namespace)
	if err != nil {
		return nil, err
	}

	if err := addAndUpdateChartRepos(helmClient, repository); err != nil {
		return nil, err
	}

	release, err := helmClient.GetRelease(chartSpec.ReleaseName)
	if err != nil {
		return nil, err
	}

	chart, _, err := helmClient.GetChart("aiida/aiida", &action.ChartPathOptions{})
	if err != nil {
		return nil, err
	}

	versionInfo := &VersionInfo{
		ReleaseName:    release.Name,
		ReleaseInfo:    release.Info,
		InstalledChart: release.Chart.Metadata,
		AvailableChart: chart.Metadata,
	}

	return versionInfo, nil
}

func Install(repository repo.Entry, chartSpec helm.ChartSpec) error {
	helmClient, err := getHelmClient(chartSpec.Namespace)
	if err != nil {
		return err
	}

	if err := addAndUpdateChartRepos(helmClient, repository); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	installedHelmChart, err := helmClient.InstallChart(ctx, &chartSpec, nil)
	if err != nil {
		return err
	}

	log.Printf("Status of Chart Install %v,\n", *installedHelmChart.Info)

	return nil
}

func Upgrade(repository repo.Entry, chartSpec helm.ChartSpec) error {
	helmClient, err := getHelmClient(chartSpec.Namespace)
	if err != nil {
		return err
	}

	if err := addAndUpdateChartRepos(helmClient, repository); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	installedHelmChart, err := helmClient.UpgradeChart(ctx, &chartSpec, nil)
	if err != nil {
		return err
	}

	log.Printf("Status of Chart Upgrade %v,\n", *installedHelmChart.Info)

	return nil
}

func Uninstall(chartSpec helm.ChartSpec) error {
	helmClient, err := getHelmClient(chartSpec.Namespace)
	if err != nil {
		return err
	}

	return helmClient.UninstallRelease(&chartSpec)
}

func addAndUpdateChartRepos(helmClient helm.Client, repository repo.Entry) error {
	if err := helmClient.AddOrUpdateChartRepo(repository); err != nil {
		return err
	}

	return helmClient.UpdateChartRepos()
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
