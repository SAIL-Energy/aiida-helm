package installer

import (
	"github.com/eddie-energy/aiida-helm/aiida-installer/internal/secrets"
	helmclient "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/repo"
	"os"
)

const AiidaReleaseName = "aiida"
const AiidaChartName = "aiida-chart"
const AiidaNamespace = "hardening"

const SealedSecretsController = "sealed-secrets-controller"

const SecretsPath = "configs/secrets-template.yaml"
const SealedSecretsPath = "configs/secrets.yaml"
const CustomAiidaValues = "configs/aiida-values.yaml"

type AiidaInstaller struct {
	Repository repo.Entry
	ChartSpec  helmclient.ChartSpec
}

func NewAiidaInstaller(repository repo.Entry) *AiidaInstaller {
	valuesOptions := values.Options{ValueFiles: []string{SealedSecretsPath}}
	if _, err := os.Stat(CustomAiidaValues); err == nil {
		valuesOptions.ValueFiles = append(valuesOptions.ValueFiles, CustomAiidaValues)
	}

	return &AiidaInstaller{
		Repository: repository,
		ChartSpec:  GetChartSpec(AiidaNamespace, repository, AiidaChartName, AiidaReleaseName, valuesOptions),
	}
}

func (aiidaInstaller *AiidaInstaller) Install() error {
	err := secrets.EncryptSecrets(AiidaNamespace, SealedSecretsController, AiidaReleaseName, SecretsPath, SealedSecretsPath)
	if err != nil {
		return err
	}

	_ = Uninstall(aiidaInstaller.ChartSpec)
	return Install(aiidaInstaller.Repository, aiidaInstaller.ChartSpec)
}
