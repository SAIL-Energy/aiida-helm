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

install:
	helm install $(RELEASE_NAME) -n $(NAMESPACE) \
	--values ./aiida-chart/values.yaml \
	--values ./aiida-chart/values-$(NAMESPACE).yaml \
	--set core.replicaCount=0 \
	--set keycloak.replicaCount=0 \
	--set nanomq.replicaCount=0 \
	--set timescaledb.replicaCount=1 \
	--set teleinfo2mqtt.replicaCount=0 \
	./aiida-chart

remove:
	helm uninstall -n $(NAMESPACE) $(RELEASE_NAME)

standby:
	helm upgrade -n $(NAMESPACE) $(RELEASE_NAME) \
	--values ./aiida-chart/values.yaml \
	--values ./aiida-chart/values-$(NAMESPACE).yaml \
	--set core.replicaCount=0 \
	--set keycloak.replicaCount=0 \
	--set nanomq.replicaCount=0 \
	--set timescaledb.replicaCount=0 \
	--set teleinfo2mqtt.replicaCount=0 \
	./aiida-chart

upgrade:
	helm upgrade -n $(NAMESPACE) $(RELEASE_NAME) \
	--values ./aiida-chart/values.yaml \
	--values ./aiida-chart/values-$(NAMESPACE).yaml \
	./aiida-chart

timescaledb-forward:
	kubectl port-forward -n $(NAMESPACE) svc/$(RELEASE_NAME)-timescaledb 15432:5432

aiida-public-forward:
	kubectl port-forward -n $(NAMESPACE) svc/$(RELEASE_NAME)-core 8333:8080

create-ghcr-secret:
	kubectl -n $(NAMESPACE) create secret docker-registry ghcr-secret \
	  --docker-server=ghcr.io \
	  --docker-username=${CR_USERNAME} \
	  --docker-password=${CR_PAT}

seal-secret:
	yq eval '.$(VALUE)' ./aiida-chart/values-$(NAMESPACE).yaml | tr -d '\n'| kubeseal \
			--controller-namespace $(NAMESPACE) \
			--raw \
			--namespace $(NAMESPACE) \
			--name $(RELEASE_NAME)-$(SECRET_NAME) | \
			xargs \-I{} yq eval -i '.$(VALUE) = "{}"' \
			./aiida-chart/values-$(NAMESPACE).yaml