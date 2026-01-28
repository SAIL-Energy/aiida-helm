CHART_PATH?=./chart
RELEASE_NAME?=aiida
NAMESPACE?=aiida
TRACK?=beta

package:
	helm package $(CHART_PATH) \
	--destination ./repository/$(TRACK)/packages

index:
	helm repo index ./repository/$(TRACK) \
	--url "https://eddie-energy.github.io/aiida-helm/$(TRACK)"

install:
	helm install $(RELEASE_NAME) -n $(NAMESPACE) \
	--values $(CHART_PATH)/values.yaml \
	--values $(CHART_PATH)/secrets.yaml \
	$(CHART_PATH)

template:
	helm template -n $(NAMESPACE) $(RELEASE_NAME) \
	--values $(CHART_PATH)/values.yaml \
	$(CHART_PATH)

upgrade:
	helm upgrade -n $(NAMESPACE) $(RELEASE_NAME) \
	--values $(CHART_PATH)/values.yaml \
	--values $(CHART_PATH)/secrets.yaml \
	$(CHART_PATH)

uninstall:
	helm uninstall -n $(NAMESPACE) $(RELEASE_NAME)

delete-all-pvcs:
	kubectl delete pvc --all -n $(NAMESPACE)