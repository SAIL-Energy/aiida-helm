RELEASE_NAME?=aiida
NAMESPACE?=hardening

template:
	helm template ./aiida-chart \
	--values ./aiida-chart/values.yaml \
	--values ./aiida-chart/values-$(NAMESPACE).yaml \
	--debug --output-dir ./out

template-plain-chart:
	helm template ./aiida-chart \
	--values ./aiida-chart/values.yaml \
	--debug --output-dir ./out

package:
	helm package ./aiida-chart \
	--destination ./aiida-repository/$(NAMESPACE)/packages

index:
	helm repo index ./aiida-repository/$(NAMESPACE) \
	--url "https://eddie-energy.github.io/aiida-helm/$(NAMESPACE)"

install:
	helm install $(RELEASE_NAME) -n $(NAMESPACE) \
	--values ./aiida-chart/values.yaml \
	--values ./aiida-chart/values-$(NAMESPACE).yaml \
	./aiida-chart

remove:
	helm uninstall -n $(NAMESPACE) $(RELEASE_NAME)

upgrade:
	helm upgrade -n $(NAMESPACE) $(RELEASE_NAME) \
	--values ./aiida-chart/values.yaml \
	--values ./aiida-chart/values-$(NAMESPACE).yaml \
	./aiida-chart

create-ghcr-secret:
	kubectl -n $(NAMESPACE) create secret docker-registry ghcr-secret \
	  --docker-server=ghcr.io \
	  --docker-username=${GHCR_USERNAME} \
	  --docker-password=${GHCR_PAT}

seal-secret:
	yq eval '.$(VALUE)' ./aiida-chart/values-$(NAMESPACE).yaml | tr -d '\n'| kubeseal \
			--controller-namespace $(NAMESPACE) \
			--raw \
			--namespace $(NAMESPACE) \
			--name $(RELEASE_NAME)-$(SECRET_NAME) | \
			xargs \-I{} yq eval -i '.$(VALUE) = "{}"' \
			./aiida-chart/values-$(NAMESPACE).yaml