CHART_PATH?=./installer/chart
RELEASE_NAME?=aiida-installer
NAMESPACE?=aiida
TRACK?=beta

create-ghcr-secret:
	kubectl -n $(NAMESPACE) create secret docker-registry ghcr-secret \
	  --docker-server=ghcr.io \
	  --docker-username=${GHCR_USERNAME} \
	  --docker-password=${GHCR_PAT}

package:
	helm package $(CHART_PATH) \
	--destination ./repository/$(TRACK)/packages

index:
	helm repo index ./repository/$(TRACK) \
	--url "https://eddie-energy.github.io/aiida-helm/$(TRACK)"

install:
	helm install $(RELEASE_NAME) -n $(NAMESPACE) \
	--values $(CHART_PATH)/values.yaml \
	$(CHART_PATH)

template:
	helm template -n $(NAMESPACE) $(RELEASE_NAME) \
	--values $(CHART_PATH)/values.yaml \
	$(CHART_PATH)

upgrade:
	helm upgrade -n $(NAMESPACE) $(RELEASE_NAME) \
	--values $(CHART_PATH)/values.yaml \
	$(CHART_PATH)

uninstall:
	helm uninstall -n $(NAMESPACE) $(RELEASE_NAME)
