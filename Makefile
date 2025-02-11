RELEASE_NAME?=aiida
NAMESPACE?=beta

template:
	helm template ./chart \
	--values ./chart/values.yaml \
	--values ./chart/values-$(NAMESPACE).yaml \
	--debug --output-dir ./out

template-plain-chart:
	helm template ./chart \
	--values ./chart/values.yaml \
	--debug --output-dir ./out

package:
	helm package ./chart \
	--destination ./repository/$(NAMESPACE)/packages

index:
	helm repo index ./repository/$(NAMESPACE) \
	--url "https://eddie-energy.github.io/aiida-helm/$(NAMESPACE)"

install:
	helm install $(RELEASE_NAME) -n $(NAMESPACE) \
	--values ./chart/values.yaml \
	--values ./chart/values-$(NAMESPACE).yaml \
	./chart

remove:
	helm uninstall -n $(NAMESPACE) $(RELEASE_NAME)

upgrade:
	helm upgrade -n $(NAMESPACE) $(RELEASE_NAME) \
	--values ./chart/values.yaml \
	--values ./chart/values-$(NAMESPACE).yaml \
	./chart

create-ghcr-secret:
	kubectl -n $(NAMESPACE) create secret docker-registry ghcr-secret \
	  --docker-server=ghcr.io \
	  --docker-username=${GHCR_USERNAME} \
	  --docker-password=${GHCR_PAT}

seal-secret:
	yq eval '.$(VALUE)' ./chart/values-$(NAMESPACE).yaml | tr -d '\n'| kubeseal \
			--controller-namespace $(NAMESPACE) \
			--raw \
			--namespace $(NAMESPACE) \
			--name $(RELEASE_NAME)-$(SECRET_NAME) | \
			xargs \-I{} yq eval -i '.$(VALUE) = "{}"' \
			./chart/values-$(NAMESPACE).yaml