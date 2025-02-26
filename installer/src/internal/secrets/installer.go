package secrets

import (
	"github.com/eddie-energy/aiida-helm/installer/internal/dto"
	helmInstaller "github.com/eddie-energy/aiida-helm/installer/internal/helm"
	helm "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/repo"
)

const ControllerName = "aiida-sealed-secrets-controller"

const repositoryUrl = "https://bitnami-labs.github.io/sealed-secrets"
const repositoryName = "sealed-secrets"
const chartName = "sealed-secrets"
const clusterRoleName = "aiida-secrets-unsealer"

type Installer struct {
	Repository repo.Entry
	ChartSpec  helm.ChartSpec
}

func NewInstaller(namespace string) *Installer {
	repository := repo.Entry{
		Name:               repositoryName,
		URL:                repositoryUrl,
		PassCredentialsAll: true,
	}
	valuesOptions := values.Options{
		StringValues: []string{
			"fullnameOverride=" + ControllerName,
			"rbac.clusterRoleName=" + clusterRoleName,
		}}

	return &Installer{
		Repository: repository,
		ChartSpec:  helmInstaller.GetChartSpec(namespace, repository, chartName, chartName, valuesOptions, true),
	}
}

func (secretsInstaller *Installer) InstallOrUpgrade() (*dto.VersionInfoDto, error) {
	return helmInstaller.InstallOrUpgrade(secretsInstaller.Repository, secretsInstaller.ChartSpec)
}

func (secretsInstaller *Installer) Uninstall() error {
	return helmInstaller.Uninstall(secretsInstaller.ChartSpec)
}
