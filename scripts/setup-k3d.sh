#!/bin/bash
set -e

rm -rf /tmp/k3dvol
mkdir /tmp/k3dvol
k3d cluster create factory --image rancher/k3s:v1.20.15-k3s1 --volume  /tmp/k3dvol:/tmp/k3dvol --registry-create local-factory-registry -p "30001:30001@loadbalancer"
kubectl create ns subscriptions
kubectl config set-context --current --namespace=subscriptions
kubectl label namespace subscriptions istio-injection=enabled
