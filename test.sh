#!/bin/zsh

NAME="py"

docker build -t rac.registry:5000/$NAME -f Dockerfile.test .

# docker pull ubuntu
# docker image tag ubuntu rac.registry:5000/myfirstimage

docker push rac.registry:5000/$NAME
