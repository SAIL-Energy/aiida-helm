package main

import (
	"github.com/eddie-energy/aiida-helm/aiida-installer/internal/installer"
	"github.com/eddie-energy/aiida-helm/aiida-installer/internal/secrets"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/repo"
	"log"
)

const RepositoryUrl = "https://eddie-energy.github.io/aiida-helm/hardening"
const RepositoryName = "aiida-repository"

const ReleaseName = "aiida"
const ChartName = "aiida-chart"
const Namespace = "hardening"

const SealedSecretsController = "sealed-secrets-controller"

const SecretsPath = "configs/secrets-template.yaml"
const SealedSecretsPath = "configs/secrets.yaml"

func main() {
	err := secrets.EncryptSecrets(Namespace, SealedSecretsController, ReleaseName, SecretsPath, SealedSecretsPath)
	if err != nil {
		log.Fatalf("Error encrypting secrets: %v", err)
	}

	aiidaRepository := repo.Entry{
		Name:               RepositoryName,
		URL:                RepositoryUrl,
		PassCredentialsAll: true,
	}

	chartSpec := installer.GetChartSpec(Namespace, aiidaRepository, ChartName, ReleaseName, values.Options{ValueFiles: []string{SealedSecretsPath}})

	err = installer.Uninstall(chartSpec)
	if err != nil {
		log.Fatalf("Error uninstalling chart: %v", err)
	}

	err = installer.Install(aiidaRepository, chartSpec)
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
- Add possibility for additional values files
*/
