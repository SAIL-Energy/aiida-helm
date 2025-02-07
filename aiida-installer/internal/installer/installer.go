package installer

import (
	"context"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/repo"
	"log"
	"time"
)

const RepositoryCache = "/tmp/.helmcache"
const RepositoryConfig = "/tmp/.helmrepo"

func GetChartSpec(namespace string, repository repo.Entry, chartName string, releaseName string, valuesOptions values.Options) helmclient.ChartSpec {
	return helmclient.ChartSpec{
		ReleaseName:     releaseName,
		ChartName:       repository.Name + "/" + chartName,
		Namespace:       namespace,
		CreateNamespace: true,
		ValuesOptions:   valuesOptions,
		SkipCRDs:        true,
		Wait:            true,
	}
}

func Install(repository repo.Entry, chartSpec helmclient.ChartSpec) error {
	myHelmClient, err := getHelmClient(chartSpec.Namespace)
	if err != nil {
		return err
	}

	if err := myHelmClient.AddOrUpdateChartRepo(repository); err != nil {
		return err
	}

	if err := myHelmClient.UpdateChartRepos(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	installedHelmChart, err := myHelmClient.InstallChart(ctx, &chartSpec, nil)
	if err != nil {
		return err
	}

	log.Printf("Status of Chart Install %v,\n", *installedHelmChart.Info)

	return nil
}

func Uninstall(chartSpec helmclient.ChartSpec) error {
	myHelmClient, err := getHelmClient(chartSpec.Namespace)
	if err != nil {
		return err
	}

	return myHelmClient.UninstallRelease(&chartSpec)
}

func getHelmClient(namespace string) (helmclient.Client, error) {
	helmOptions := &helmclient.Options{
		Namespace:        namespace,
		RepositoryCache:  RepositoryCache,
		RepositoryConfig: RepositoryConfig,
		Debug:            true,
		Linting:          true,
		DebugLog:         func(format string, v ...interface{}) { log.Printf(format+"\n", v...) },
	}

	return helmclient.New(helmOptions)
}
