#!/bin/sh

set -e

if docker ps | grep rac.registry; then
  docker rm -f rac.registry
fi

echo "building registry"
docker build -t rac.registry:local -f Dockerfile.registry .

until docker network ls | grep k3d-k3s-default; do
  echo "waiting for network to create"
done

docker run -d --rm --name rac.registry -p 5000:5000 --network k3d-k3s-default rac.registry:local
grep -qxF "127.0.0.1 rac.registry" /etc/hosts || $(echo "need to add registry to etc/hosts" && sudo echo "127.0.0.1 rac.registry" >> /etc/hosts)
