

# RAC (Run A Container)
## Goal
 - Make it easy to deploy a Docker container to the edge
 - Simplify dev's life
 - Container spins up when request comes in on URL, close as possible to the requester
## Details
 - 2 entrypoints / major components: Build repo -> image, and Deploy image -> container
 - No sign up required for Deploy -- i.e you can push a container to our registry, and it will deploy
 - Sign up unlocks Build step, and further config around your deployment
 - Anyone can push to any repo on RAC registry, unless its been claimed by a registered user
	 - means if you don't sign up, you don't have guaranteed control over your repo monkaS
## Deploy
 - Run a Docker Registry w/ [token auth](https://docs.docker.com/registry/spec/auth/token/), [S3 storage](https://docs.docker.com/registry/storage-drivers/s3/), and [notifications](https://docs.docker.com/registry/notifications/)
 - Token auth gives all the auth power to RAC API, which prevents logged out users from pushing to claimed repositories (and allows users to login to push to their repos)
 - S3 storage idk its just there, might be too slow or smth I haven't really tried it
 - Notifications tell the API everytime someone pushes a new repository to RAC registry
 - Notifications interfaces w/ a K3s cluster to create pod/deployment/service or whatever. I'm a k8s noob so don't have this part fully fleshed out yet. Rn I have it creating a pod and deployment, but we obv don't want to do that everytime a new push comes in, only when its requested (I think)
	 - not sure how we do the "when a request comes in, spin up container on edge" part. run nginx/some proxy on edge nodes that are all part of k3s cluster, and when a request comes in tell API and make sure there's a container running?
	 - or is there some k8s tooling that would be great here?
- There should be **no config file**. Dockerfile can expose ports and volumes, which is the basic information. For more advanced config we can use [Docker Labels](https://docs.docker.com/config/labels-custom-metadata/) 

## Build
 - Haven't spent as much time on this yet. I think getting the deploy part working first is more important, cuz you can have an MVP then
 - Building on [eggroll](https://github.com/zzh8829/eggroll) should have Github auth, you can onboard certain Github repos
 - Whenever a commit is pushed (to master?) on one of those repos, build the commit into a Docker image and push it to RAC Registry

# Current State
## Deploy
- Run `./run.sh` in this monorepo to spin up and connect:
	- a [k3d](https://docs.docker.com/config/labels-custom-metadata/) cluster
	- a docker registry (doesn't have lasting storage rn, makes it easier to test)
	- RAC API which handles auth and notifications
- K3D 
	- 2 docker containers -- 1 agent 1 master
- registry
	- configured to talk to RAC API for notifications and auth and use proper certs and shit
- RAC API
	- Gin HTTP server
	- has auth and notifications route
	- doesn't have real auth, rn there's hardcoded users and docker clients are allowed to do basically anything
	- has a /claim route that accepts any user_id as claiming any repo that hasn't yet been claimed
- `./test.sh` builds a basic image and pushes it to local registry -- should cause a deployment on the local K3D cluster
## Build
- nothing m8

# Useful Commands
## JWT
The registry and API use an x509 cert to sign JWTs, use this command to generate a new cert & private key:
```bash
openssl req -new -newkey rsa:1024 -days 365 -nodes -x509 -keyout jwt.key -out jwt.cert
```
