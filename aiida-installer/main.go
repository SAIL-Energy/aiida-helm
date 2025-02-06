package main

import (
	"bytes"
	"context"
	"github.com/eddie-energy/aiida-helm/aiida-installer/internal/secrets"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/repo"
	"log"
	"time"
)

const RepositoryUrl = "https://eddie-energy.github.io/aiida-helm/hardening"
const RepositoryName = "aiida-repository"
const RepositoryCache = "/tmp/.helmcache"
const RepositoryConfig = "/tmp/.helmrepo"

const ReleaseName = "aiida"
const ChartName = "aiida-chart"
const Namespace = "hardening"

const SealedSecretsController = "sealed-secrets-controller"

const SecretsPath = "configs/secrets-template.yaml"
const SealedSecretsPath = "configs/secrets.yaml"

func install(namespace string) error {
	var outputBuffer bytes.Buffer

	opt := &helmclient.Options{
		Namespace:        namespace,
		RepositoryCache:  RepositoryCache,
		RepositoryConfig: RepositoryConfig,
		Debug:            true,
		Linting:          true,
		DebugLog:         func(format string, v ...interface{}) { log.Printf(format+"\n", v...) },
		Output:           &outputBuffer,
	}

	myHelmClient, err := helmclient.New(opt)
	if err != nil {
		return err
	}

	repository := repo.Entry{
		Name:               RepositoryName,
		URL:                RepositoryUrl,
		PassCredentialsAll: true,
	}

	if err := myHelmClient.AddOrUpdateChartRepo(repository); err != nil {
		return err
	}

	if err := myHelmClient.UpdateChartRepos(); err != nil {
		return err
	}

	chartSpec := helmclient.ChartSpec{
		ReleaseName:     ReleaseName,
		ChartName:       RepositoryName + "/" + ChartName,
		Namespace:       namespace,
		CreateNamespace: true,
		ValuesOptions:   values.Options{ValueFiles: []string{SealedSecretsPath}},
		SkipCRDs:        true,
		Wait:            true,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	_ = myHelmClient.UninstallRelease(&chartSpec)

	myInstalledHelmChart, err := myHelmClient.InstallChart(ctx, &chartSpec, nil)
	if err != nil {
		return err
	}

	log.Printf("Status of Chart Install %v,\n", *myInstalledHelmChart.Info)

	return nil
}

func main() {
	err := secrets.EncryptSecrets(Namespace, SealedSecretsController, ReleaseName, SecretsPath, SealedSecretsPath)
	if err != nil {
		log.Fatalf("Error encrypting secrets: %v", err)
	}

	err = install(Namespace)
	if err != nil {
		log.Fatalf("Error installing chart: %v", err)
	}
}

/*
TODO
- Error handling, logging, tests
- Remove, upgrade, install AIIDA + sealed secrets controller
- Remove, upgrade, install other services
- Add REST access from AIIDA core (restricted access!)
- Restrict access to namespace
*/
