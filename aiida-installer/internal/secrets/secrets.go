package secrets

import (
	"bytes"
	"context"
	"crypto/rsa"
	"fmt"
	ssv1alpha1 "github.com/bitnami-labs/sealed-secrets/pkg/apis/sealedsecrets/v1alpha1"
	"github.com/bitnami-labs/sealed-secrets/pkg/kubeseal"
	"github.com/sethvargo/go-password/password"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

const secretNameSuffix = "secret"

func EncryptSecrets(namespace string, controllerName string, releaseName string, secretsPath string, sealedSecretsPath string) error {
	secrets, err := readFromYamlFile(secretsPath)
	if err != nil {
		return err
	}

	pubKey, err := GetPublicKey(namespace, controllerName)
	if err != nil {
		return err
	}

	for deploymentName, deploymentSecrets := range secrets {
		for secretKey, secretItem := range deploymentSecrets {
			secretName := fmt.Sprintf("%s-%s-%s", releaseName, deploymentName, secretNameSuffix)

			if secretItem == "" {
				secretItem, err = password.Generate(64, 10, 10, false, false)

				if err != nil {
					return err
				}
			}

			encryptedSecretItem, err := EncryptSecret(namespace, pubKey, secretName, secretItem)
			if err != nil {
				return err
			}

			secrets[deploymentName][secretKey] = *encryptedSecretItem
		}
	}

	return writeToYamlFile(sealedSecretsPath, secrets)
}

func EncryptSecret(namespace string, pubKey *rsa.PublicKey, secretName string, secretItem string) (*string, error) {
	encryptedSecretItemBuffer := new(bytes.Buffer)

	err := kubeseal.EncryptSecretItem(
		encryptedSecretItemBuffer,
		secretName,
		namespace,
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

func GetPublicKey(namespace string, controllerName string) (*rsa.PublicKey, error) {
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	ctx := context.Background()

	certFile, err := kubeseal.OpenCert(ctx, clientConfig, namespace, controllerName, "")
	if err != nil {
		return nil, err
	}

	pubKey, err := kubeseal.ParseKey(certFile)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

func readFromYamlFile(filePath string) (map[string]map[string]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var values map[string]map[string]string
	err = yaml.Unmarshal(data, &values)
	if err != nil {
		return nil, err
	}

	return values, nil
}

func writeToYamlFile(filePath string, values map[string]map[string]string) error {
	encryptedValues, err := yaml.Marshal(values)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, encryptedValues, 0644)
}
