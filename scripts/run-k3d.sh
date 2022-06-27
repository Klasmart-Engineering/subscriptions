#!/bin/bash
set -e

cluster=$(kubectl config current-context)

if [[ $cluster != *"k3d"* ]]; then
  echo "OOPS! trying to deploy to another cluster"
  exit 1
fi

kubectl config set-context --current --namespace=subscriptions
tilt up -f ./environment/local/Tiltfile
