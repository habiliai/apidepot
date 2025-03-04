#!/bin/bash

set -e

cd $(dirname $0)

cat << EOF > docker_images.txt
  busybox
  postgres:14
  traefik:v2.10
  postgres:14-alpine
  traefik/whoami:latest
  supabase/storage-api:v1.0.10
  quay.io/minio/mc:latest
  quay.io/minio/minio:latest
EOF

export DOCKER_IMAGES=$(cat docker_images.txt)

for IMG in $DOCKER_IMAGES; do
  docker pull $IMG
done