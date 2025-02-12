package aiida

import (
	helmInstaller "github.com/eddie-energy/aiida-helm/installer/internal/helm"
	"github.com/eddie-energy/aiida-helm/installer/internal/secrets"
	helm "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/repo"
	"os"
)

const chartName = "aiida"
const namespace = "hardening"

const secretsPath = "configs/secrets-template.yaml"
const sealedSecretsPath = "configs/secrets.yaml"
const customValues = "configs/aiida-values.yaml"

type Installer struct {
	Repository       repo.Entry
	ChartSpec        helm.ChartSpec
	SecretsInstaller *secrets.Installer
	Encryptor        *secrets.Encryptor
}

func NewInstaller(repository repo.Entry) *Installer {
	valuesOptions := values.Options{ValueFiles: []string{sealedSecretsPath}}
	if _, err := os.Stat(customValues); err == nil {
		valuesOptions.ValueFiles = append(valuesOptions.ValueFiles, customValues)
	}

	return &Installer{
		Repository:       repository,
		ChartSpec:        helmInstaller.GetChartSpec(namespace, repository, chartName, valuesOptions, false),
		SecretsInstaller: secrets.NewInstaller(namespace),
		Encryptor:        secrets.NewEncryptor(namespace, secrets.ControllerName, chartName, secretsPath, sealedSecretsPath),
	}
}

func (aiidaInstaller *Installer) GetVersionInfo() (*helmInstaller.VersionInfo, error) {
	return helmInstaller.GetVersionInfo(aiidaInstaller.Repository, aiidaInstaller.ChartSpec)
}

func (aiidaInstaller *Installer) InstallOrUpgrade() (*helmInstaller.VersionInfo, error) {
	if _, err := aiidaInstaller.SecretsInstaller.InstallOrUpgrade(); err != nil {
		return nil, err
	}

	err := aiidaInstaller.Encryptor.EncryptSecrets()
	if err != nil {
		return nil, err
	}

	return helmInstaller.InstallOrUpgrade(aiidaInstaller.Repository, aiidaInstaller.ChartSpec)
}

func (aiidaInstaller *Installer) Uninstall() error {
	err := helmInstaller.Uninstall(aiidaInstaller.ChartSpec)
	_ = aiidaInstaller.SecretsInstaller.Uninstall()

	return err
}
