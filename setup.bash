#!/usr/bin/bash

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