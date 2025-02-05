package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	helmclient "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/repo"
)

func main() {
	var outputBuffer bytes.Buffer

	opt := &helmclient.Options{
		Namespace:        "hardening",
		RepositoryCache:  "/tmp/.helmcache",
		RepositoryConfig: "/tmp/.helmrepo",
		Debug:            false,
		Linting:          true,
		DebugLog:         func(format string, v ...interface{}) {},
		Output:           &outputBuffer,
	}

	myHelmClient, err := helmclient.New(opt)
	if err != nil {
		panic(err)
	}

	chartRepo := repo.Entry{
		Name:               "aiida-repo",
		URL:                "https://eddie-energy.github.io/aiida-helm/hardening",
		PassCredentialsAll: true,
	}

	if err := myHelmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("Added Chart Repo %s\n", chartRepo.Name)
	}

	if err := myHelmClient.UpdateChartRepos(); err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("Updating Chart Repo\n")
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName:     "aiida-chart",
		ChartName:       "aiida-repo/aiida-chart",
		Namespace:       "hardening",
		CreateNamespace: true,
		ValuesOptions: values.Options{
			StringValues: []string{
				"controller.status.processors=20",
				"controller.operation.processors=10",
				"controller.self.heal.timeout.seconds=5",
				"controller.repo.server.timeout.seconds=60",
			},
			Values: []string{"./values-dev.yaml"},
		},
		SkipCRDs: true,
		Wait:     true,
	}
	ctx := context.Background()
	mycontext, _ := context.WithTimeout(ctx, 80000*time.Second)
	myInstalledHelmChart, err := myHelmClient.InstallChart(mycontext, &chartSpec, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Status of Chart Install %v,\n", *myInstalledHelmChart.Info)
}
