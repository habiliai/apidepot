#!/bin/bash

kubectl --context kind-apidepot apply -f https://raw.githubusercontent.com/metallb/metallb/v0.13.7/config/manifests/metallb-native.yaml
kubectl --context kind-apidepot get pod -n metallb-system
kubectl --context kind-apidepot wait --namespace metallb-system \
        --for=condition=ready pod \
        --selector=app=metallb \
        --timeout=90s