#!/bin/sh

set -e

if k3d ls | grep k3s-default; then
  k3d d
fi

echo "starting k3d cluster"
k3d c --workers 1 --registry-name rac.registry -v `pwd`/registries.yml:/etc/rancher/k3s/registries.yaml --server-arg '--tls-san=k3d-k3ds-default-server'
