package aiida

import (
	"github.com/eddie-energy/aiida-helm/installer/internal/cluster"
	"github.com/eddie-energy/aiida-helm/installer/internal/dto"
	helmInstaller "github.com/eddie-energy/aiida-helm/installer/internal/helm"
	"github.com/eddie-energy/aiida-helm/installer/internal/secrets"
	helm "github.com/mittwald/go-helm-client"
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/repo"
	"log"
	"os"
	"strings"
	"time"
)

const ChartName = "aiida"
const namespace = "aiida"

const customValues = "configs/aiida-values.yaml"
const secretsPath = "configs/secrets-template.yaml"
const sealedSecretsPath = "data/secrets.yaml"

const hostPrefix = "ingress.host="
const ipAddressFetchInterval = 30 * time.Second

type Installer struct {
	Repository       repo.Entry
	ChartSpec        helm.ChartSpec
	SecretsInstaller *secrets.Installer
	Encryptor        *secrets.Encryptor
	ClusterClient    *cluster.Client
}

func NewInstaller(repository repo.Entry) *Installer {
	valuesOptions := values.Options{ValueFiles: []string{sealedSecretsPath}}
	if _, err := os.Stat(customValues); err == nil {
		valuesOptions.ValueFiles = append(valuesOptions.ValueFiles, customValues)
	}

	aiidaInstaller := &Installer{
		Repository:       repository,
		ChartSpec:        helmInstaller.GetChartSpec(namespace, repository, ChartName, ChartName, valuesOptions, false),
		SecretsInstaller: secrets.NewInstaller(namespace),
		Encryptor:        secrets.NewEncryptor(namespace, secrets.ControllerName, ChartName, secretsPath, sealedSecretsPath),
		ClusterClient:    &cluster.Client{},
	}

	overrideHostWithIp := os.Getenv("OVERRIDE_HOST_WITH_IP")
	if overrideHostWithIp == "true" {
		address := *aiidaInstaller.updateIpAddress()
		log.Println("Initial address:", address)

		go aiidaInstaller.startIpAddressUpdater(address)
	}

	versionInfo, err := aiidaInstaller.InstallOrUpgrade()
	if err != nil {
		log.Fatalf("Failed to install or upgrade aiida: %v", err.Error())
	}
	log.Printf("Installed or upgraded aiida: %v", versionInfo.InstalledChart.Version)

	return aiidaInstaller
}

func (aiidaInstaller *Installer) GetVersionInfo() (*dto.VersionInfoDto, error) {
	return helmInstaller.GetVersionInfo(aiidaInstaller.Repository, aiidaInstaller.ChartSpec)
}

func (aiidaInstaller *Installer) InstallOrUpgrade() (*dto.VersionInfoDto, error) {
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

func (aiidaInstaller *Installer) updateIpAddress() *string {
	address, err := aiidaInstaller.ClusterClient.GetPrimaryServerAddress()
	if err != nil {
		log.Println("Failed to get primary server address:", err)
		return nil
	}

	if address != nil {
		hostValue := hostPrefix + *address
		replaced := false

		for i, val := range aiidaInstaller.ChartSpec.ValuesOptions.StringValues {
			if strings.HasPrefix(val, hostPrefix) {
				aiidaInstaller.ChartSpec.ValuesOptions.StringValues[i] = hostValue
				replaced = true
				break
			}
		}

		if !replaced {
			aiidaInstaller.ChartSpec.ValuesOptions.StringValues = append(aiidaInstaller.ChartSpec.ValuesOptions.StringValues, hostValue)
		}
	}

	return address
}

func (aiidaInstaller *Installer) startIpAddressUpdater(initialAddress string) {
	currentAddress := initialAddress
	ticker := time.NewTicker(ipAddressFetchInterval)
	defer ticker.Stop()

	for range ticker.C {
		address := aiidaInstaller.updateIpAddress()
		if address != nil && currentAddress != *address {
			log.Println("Updated address to:", *address)
			currentAddress = *address
			_, err := aiidaInstaller.InstallOrUpgrade()
			if err != nil {
				log.Println("Failed to update aiida host address:", err)
			}
		}
	}
}
