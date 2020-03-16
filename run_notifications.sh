#!/bin/sh
set -e

if docker ps | grep rac.notifications; then
  docker rm -f rac.notifications
fi

until k3d get-kubeconfig --name='k3s-default'; do
  echo "waiting for k3d kubeconfig to exist"
  sleep 1
done

KUBECONFIG=$(k3d get-kubeconfig --name='k3s-default')

cp $KUBECONFIG ./kubeconfig
sed -i 's/localhost/k3d-k3s-default-server/g' ./kubeconfig

until kubectl get secrets | grep "default-token"; do
  echo "waiting for cert generation"
done

kubectl get secrets $(kubectl get secrets | grep "default-token" | cut -d' ' -f 1) -o go-template='{{index .data "ca.crt" | base64decode}}' > ./ca.crt

echo "building notifications"
set -x
docker build -t rac.notifications:local -f Dockerfile.notifications --build-arg KUBECONFIG=./kubeconfig --build-arg K3D_CERT=./ca.crt .


rm ./kubeconfig ./ca.crt

docker run --rm --name rac.notifications --network k3d-k3s-default rac.notifications:local
