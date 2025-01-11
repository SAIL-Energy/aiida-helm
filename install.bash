#!/usr/bin/bash

curl -sfL https://get.k3s.io | sh -
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

sudo pacman -S kubectl
sudo pacman -S helm
sudo pacman -S k9s
sudo pacman -S yq
sudo pacman -S kubeseal