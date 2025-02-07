package aiida

import (
	helmInstaller "github.com/eddie-energy/aiida-helm/aiida-installer/internal/helm"
	"github.com/eddie-energy/aiida-helm/aiida-installer/internal/secrets"
	helm "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/repo"
	"os"
)

const chartName = "aiida"
const namespace = "hardening"

const sealedSecretsController = "sealed-secrets-controller"

const secretsPath = "configs/secrets-template.yaml"
const sealedSecretsPath = "configs/secrets.yaml"
const customValues = "configs/aiida-values.yaml"

type Installer struct {
	Repository repo.Entry
	ChartSpec  helm.ChartSpec
}

func NewInstaller(repository repo.Entry) *Installer {
	valuesOptions := values.Options{ValueFiles: []string{sealedSecretsPath}}
	if _, err := os.Stat(customValues); err == nil {
		valuesOptions.ValueFiles = append(valuesOptions.ValueFiles, customValues)
	}

	return &Installer{
		Repository: repository,
		ChartSpec:  helmInstaller.GetChartSpec(namespace, repository, chartName, valuesOptions),
	}
}

func (aiidaInstaller *Installer) GetVersionInfo() (*helmInstaller.VersionInfo, error) {
	return helmInstaller.GetSingleVersionInfo(aiidaInstaller.Repository, aiidaInstaller.ChartSpec)
}

func (aiidaInstaller *Installer) Install() error {
	if err := secrets.EncryptSecrets(namespace, sealedSecretsController, chartName, secretsPath, sealedSecretsPath); err != nil {
		return err
	}

	return helmInstaller.Install(aiidaInstaller.Repository, aiidaInstaller.ChartSpec)
}

func (aiidaInstaller *Installer) Upgrade() error {
	return helmInstaller.Upgrade(aiidaInstaller.Repository, aiidaInstaller.ChartSpec)
}
