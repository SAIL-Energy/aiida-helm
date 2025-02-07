package installer

import (
	"github.com/mittwald/go-helm-client/values"
	"helm.sh/helm/v3/pkg/repo"
)

const ServiceReleaseNamePrefix = "aiida-service"
const ServiceNamespace = "aiida-services"

type ServiceInstaller struct {
	Repository repo.Entry
}

func NewServiceInstaller(repository repo.Entry) *ServiceInstaller {
	return &ServiceInstaller{Repository: repository}
}

func (serviceInstaller *ServiceInstaller) Install(chartName string) error {
	releaseName := ServiceReleaseNamePrefix + "-" + chartName
	chartSpec := GetChartSpec(ServiceNamespace, serviceInstaller.Repository, chartName, releaseName, values.Options{})

	_ = Uninstall(chartSpec)
	return Install(serviceInstaller.Repository, chartSpec)
}
