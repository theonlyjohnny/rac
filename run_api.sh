#!/bin/sh
set -e

if docker ps | grep rac.api; then
  docker rm -f rac.api
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

echo "building api"
set -x
docker build -t rac.api:local -f Dockerfile.api --build-arg KUBECONFIG=./kubeconfig --build-arg K3D_CERT=./ca.crt .


rm ./kubeconfig ./ca.crt

docker run --rm --name rac.api --network k3d-k3s-default -p 8090:8090 -v `pwd`/jwt.key:/var/jwt.key rac.api:local
