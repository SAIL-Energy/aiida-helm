#!/usr/bin/bash

NAMESPACE_ARG="--namespace"
NAMESPACE="hardening"
GHCR_USERNAME_ARG="--ghcr-username"
GHCR_USERNAME=""
GHCR_PAT_ARG="--ghcr-pat"
GHCR_PAT=""

display_help () {
  echo -e "This script installs the required tools to deploy AIIDA on a k3s cluster."
  echo -e "First enter secrets into the values file for the namespace you want to deploy to."
  echo -e "Usage: setup.bash [OPTIONS]"
  echo -e "Options:"
  echo -e "  --help\t\t\tDisplay this help message."
  echo -e "  --ghcr-username=<username>\tThe GitHub Container Registry username."
  echo -e "  --ghcr-pat=<pat>\t\tThe GitHub Container Registry Personal Access Token."
  echo -e "  --namespace=<namespace>\t(Optional) The namespace to deploy the services to. Default: hardening."
}

for arg in "$@"; do
  case "$arg" in
    --help)
      display_help
      exit 0
      ;;
    ${NAMESPACE_ARG}*)
      NAMESPACE="${arg#*=}"
      ;;
    ${GHCR_USERNAME_ARG}*)
      GHCR_USERNAME="${arg#*=}"
      ;;
    ${GHCR_PAT_ARG}*)
      GHCR_PAT="${arg#*=}"
      ;;
    *)
      echo "Unknown argument: $arg"
      exit 1
      ;;
  esac
done

if [ -z "$GHCR_USERNAME" ] || [ -z "$GHCR_PAT" ]; then
  echo "Missing argument(s): $GHCR_USERNAME_ARG or $GHCR_PAT_ARG"
  exit 1
fi

curl -sfL https://get.k3s.io | sh -
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
chmod 644 /etc/rancher/k3s/k3s.yaml

sudo pacman -S kubectl
sudo pacman -S helm
sudo pacman -S k9s
sudo snap install yq
sudo pacman -S kubeseal

helm repo add sealed-secrets https://bitnami-labs.github.io/sealed-secrets

helm install sealed-secrets \
     -n "$NAMESPACE" \
     --set-string fullnameOverride=sealed-secrets-controller \
     sealed-secrets/sealed-secrets

helm repo add jetstack https://charts.jetstack.io --force-update

helm install cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true \
  jetstack/cert-manager

make create-ghcr-secret

make seal-secret VALUE="timescaledb.encryptedUsername" SECRET_NAME="timescaledb-secret"
make seal-secret VALUE="timescaledb.encryptedPassword" SECRET_NAME="timescaledb-secret"

make seal-secret VALUE="nanomq.encryptedPwd" SECRET_NAME="nanomq-secret"

make seal-secret VALUE="teleinfo2mqtt.encryptedMqttUsername" SECRET_NAME="teleinfo2mqtt-secret"
make seal-secret VALUE="teleinfo2mqtt.encryptedMqttPassword" SECRET_NAME="teleinfo2mqtt-secret"

make seal-secret VALUE="keycloak.encryptedAdminUsername" SECRET_NAME="keycloak-secret"
make seal-secret VALUE="keycloak.encryptedAdminPassword" SECRET_NAME="keycloak-secret"

make install