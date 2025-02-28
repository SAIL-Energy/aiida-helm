package secrets

import (
	"bytes"
	"context"
	"crypto/rsa"
	"fmt"
	"github.com/awnumar/memguard"
	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealedsecrets/v1alpha1"
	"github.com/bitnami-labs/sealed-secrets/pkg/kubeseal"
	"github.com/sethvargo/go-password/password"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"strings"
)

const secretNameSuffix = "secret"

const encryptionKeyPrefix = "encrypted"
const autoGenerateValue = "*"
const autoGenerateIfNotEmptyValue = "*!"

const tokenDeploymentName = "core"
const tokenSecretKey = "encryptedInstallerToken"

type Encryptor struct {
	Namespace             string
	ControllerName        string
	ReleaseName           string
	SecretsPath           string
	SealedSecretsPath     string
	InstallerTokenEnclave *memguard.Enclave
}

func NewEncryptor(
	namespace string,
	controllerName string,
	releaseName string,
	secretsPath string,
	sealedSecretsPath string,
) *Encryptor {
	return &Encryptor{
		Namespace:         namespace,
		ControllerName:    controllerName,
		ReleaseName:       releaseName,
		SecretsPath:       secretsPath,
		SealedSecretsPath: sealedSecretsPath,
	}
}

func (encryptor *Encryptor) EncryptSecrets() error {
	secrets, err := encryptor.readFromYamlFile(encryptor.SecretsPath)
	if err != nil {
		return err
	}

	var currentSecrets map[string]map[string]string
	if _, err := os.Stat(encryptor.SealedSecretsPath); err == nil {
		currentSecrets, err = encryptor.readFromYamlFile(encryptor.SealedSecretsPath)
		if err != nil {
			return err
		}
	}

	pubKey, err := encryptor.GetPublicKey()
	if err != nil {
		return err
	}

	var token string

	for deploymentName, deploymentSecrets := range secrets {
		for secretKey, secretItem := range deploymentSecrets {
			secretName := fmt.Sprintf("%s-%s-%s", encryptor.ReleaseName, deploymentName, secretNameSuffix)

			if secretItem == autoGenerateIfNotEmptyValue {
				currentSecret := currentSecrets[deploymentName][secretKey]
				if currentSecret != "" {
					secretItem = currentSecret
				} else {
					secretItem = autoGenerateValue
				}
			}

			if secretItem == autoGenerateValue {
				secretItem, err = password.Generate(32, 8, 0, false, true)

				if err != nil {
					return err
				}
			}

			if deploymentName == tokenDeploymentName && secretKey == tokenSecretKey {
				token = secretItem
			}

			if strings.HasPrefix(secretKey, encryptionKeyPrefix) {
				encryptedSecretItem, err := encryptor.EncryptSecret(pubKey, secretName, secretItem)
				if err != nil {
					return err
				}
				secretItem = *encryptedSecretItem
			}

			secrets[deploymentName][secretKey] = secretItem
		}
	}

	tokenBuffer := memguard.NewBufferFromBytes([]byte(token))
	defer tokenBuffer.Destroy()
	encryptor.InstallerTokenEnclave = tokenBuffer.Seal()

	return encryptor.writeToYamlFile(encryptor.SealedSecretsPath, secrets)
}

func (encryptor *Encryptor) EncryptSecret(pubKey *rsa.PublicKey, secretName string, secretItem string) (*string, error) {
	encryptedSecretItemBuffer := new(bytes.Buffer)

	err := kubeseal.EncryptSecretItem(
		encryptedSecretItemBuffer,
		secretName,
		encryptor.Namespace,
		[]byte(secretItem),
		ssv1alpha1.StrictScope,
		pubKey,
	)
	if err != nil {
		return nil, err
	}

	encryptedSecretItem := encryptedSecretItemBuffer.String()
	return &encryptedSecretItem, nil
}

func (encryptor *Encryptor) GetPublicKey() (*rsa.PublicKey, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	ctx := context.Background()

	certFile, err := kubeseal.OpenCert(ctx, clientConfig, encryptor.Namespace, encryptor.ControllerName, "")
	if err != nil {
		return nil, err
	}

	pubKey, err := kubeseal.ParseKey(certFile)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

func (encryptor *Encryptor) readFromYamlFile(filePath string) (map[string]map[string]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var values map[string]map[string]string

	if err = yaml.Unmarshal(data, &values); err != nil {
		return nil, err
	}

	return values, nil
}

func (encryptor *Encryptor) writeToYamlFile(filePath string, values map[string]map[string]string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	encryptedValues, err := yaml.Marshal(values)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, encryptedValues, 0644)
}
