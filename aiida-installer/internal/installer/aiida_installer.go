package installer

import (
	"github.com/eddie-energy/aiida-helm/aiida-installer/internal/secrets"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/repo"
)

const ReleaseName = "aiida"
const ChartName = "aiida-chart"
const Namespace = "hardening"

const SealedSecretsController = "sealed-secrets-controller"

const SecretsPath = "configs/secrets-template.yaml"
const SealedSecretsPath = "configs/secrets.yaml"

type AiidaInstaller struct {
	Repository repo.Entry
	ChartSpec  helmclient.ChartSpec
}

func NewAiidaInstaller(repository repo.Entry) *AiidaInstaller {
	return &AiidaInstaller{
		Repository: repository,
		ChartSpec: GetChartSpec(
			Namespace,
			repository,
			ChartName,
			ReleaseName,
			values.Options{ValueFiles: []string{SealedSecretsPath}},
		),
	}
}

func (aiidaInstaller *AiidaInstaller) Install() error {
	err := secrets.EncryptSecrets(Namespace, SealedSecretsController, ReleaseName, SecretsPath, SealedSecretsPath)
	if err != nil {
		return err
	}

	_ = Uninstall(aiidaInstaller.ChartSpec)
	return Install(aiidaInstaller.Repository, aiidaInstaller.ChartSpec)
}
