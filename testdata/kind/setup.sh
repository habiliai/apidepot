#!/bin/bash
kind create cluster --config kind.yaml
kubectl wait --context kind-apidepot --for=condition=ready node --selector=type=default --timeout=60s

kind load docker-image $DOCKER_IMAGES --name apidepot --verbosity 1

bash -e ./yamls/metallb/metallb.sh
bash -e ./yamls/traefik/traefik.sh
kubectl apply --context kind-apidepot -f "yamls/*"
kubectl rollout --context kind-apidepot -n kube-system restart deploy/coredns
kubectl wait --context kind-apidepot --namespace default --for=condition=ready --timeout=300s pod -l group=apidepot
