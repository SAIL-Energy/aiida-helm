package cluster

import (
	"context"
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type apiServerInfo struct {
	ServerAddressByClientCIDRs []struct {
		ServerAddress string `json:"serverAddress"`
	} `json:"serverAddressByClientCIDRs"`
}

type Client struct {
	serverAddress string
}

func (clusterClient Client) GetPrimaryServerAddress() (*string, error) {
	info, err := clusterClient.fetchAPIServerInfo()
	if err != nil {
		return nil, err
	}

	serverAddress := strings.Split(info.ServerAddressByClientCIDRs[0].ServerAddress, ":")[0]
	return &serverAddress, nil
}

func (clusterClient Client) fetchAPIServerInfo() (*apiServerInfo, error) {
	restClient, err := clusterClient.getRESTClient()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rawData, err := restClient.Get().RequestURI("/api").Do(ctx).Raw()
	if err != nil {
		return nil, err
	}

	var info apiServerInfo
	if err := json.Unmarshal(rawData, &info); err != nil {
		return nil, err
	}

	if len(info.ServerAddressByClientCIDRs) == 0 {
		return nil, fmt.Errorf("no server address found in response")
	}

	return &info, nil
}

func (clusterClient Client) getRESTClient() (*rest.RESTClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		homeDir, _ := os.UserHomeDir()
		kubeConfig := filepath.Join(homeDir, ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			return nil, err
		}
	}

	config.GroupVersion = &schema.GroupVersion{Version: "v1"}
	config.APIPath = "/api"
	config.NegotiatedSerializer = serializer.NewCodecFactory(nil).WithoutConversion()

	return rest.RESTClientFor(config)
}
