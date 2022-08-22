#!/bin/sh

set -e

BUILD_TYPE="dev"
BUILD_VERSION=$(git rev-list -1 --abbrev-commit HEAD)
BUILD_DATE="$(date -u)"

rm -rf vendor

dev_docker_build() {
  docker build --build-arg BUILD_DATE="$BUILD_DATE" \
                --build-arg BUILD_TYPE="$BUILD_TYPE" \
                --build-arg BUILD_VERSION="$BUILD_VERSION" \
                -t automated-approver:dev \
                -f dockerfiles/Dockerfile .
  
  rm -rf vendor
}

dev_docker_test_build() {
  docker build --build-arg BUILD_DATE="$BUILD_DATE" \
                --build-arg BUILD_TYPE="$BUILD_TYPE" \
                --build-arg BUILD_VERSION="$BUILD_VERSION" \
                -t automated-approver-unittest:dev \
                -f dockerfiles/DockerfileUnitTest .
  
  rm -rf vendor
}

dev_docker_build_multiarch() {
  # We need to build the images one by one so they can be exported (doesn't work otherwise)
  # If this command fails because of buildx, please run the following command:
  # docker buildx create --use
  for arch in amd64 arm64 ; do
      docker buildx build \
      --build-arg BUILD_DATE="$BUILD_DATE" \
      --build-arg BUILD_TYPE="$BUILD_TYPE" \
      --build-arg BUILD_VERSION="$BUILD_VERSION" \
      --platform linux/$arch \
      --output "type=docker,push=false,name=automated-approver:dev-$arch,dest=automated-approver-$arch.tar" \
      -f dockerfiles/Dockerfile .
  done

  rm -rf vendor
}


dev_local_build() {
  go mod tidy
  go build \
      -tags debug \
      -ldflags "-X 'main.buildDate=$BUILD_DATE' \
                -X 'main.buildVersion=$BUILD_VERSION' \
                -X 'main.buildType=$BUILD_TYPE'" \
      -o out/automated-approver \
      gitlab.qredo.com/custody-engine/automated-approver/cmd/service
}


go mod tidy

if [ -n "$1" ]; then
  case $1 in
    docker)
      dev_docker_build
      ;;
    docker_multiarch)
      dev_docker_build_multiarch
      ;;
    docker_unittest)
      dev_docker_test_build
      ;;
  esac
else
  dev_local_build
fi
