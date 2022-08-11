#!/bin/sh

set -e

GIT_COMMIT=$(git rev-list -1 --abbrev-commit HEAD)
BUILD_DATE="$(date -u)"
VERSION="dev"

rm -rf vendor

dev_docker_build() {

  docker build --build-arg BUILD_DATE="$BUILD_DATE" --build-arg BUILD_COMMIT="$GIT_COMMIT" --build-arg BUILD_VERSION="$VERSION" -t automated-approver:dev -f dockerfiles/Dockerfile .
  rm -rf vendor
}

dev_docker_test_build() {

  docker build --build-arg BUILD_DATE="$BUILD_DATE" --build-arg BUILD_COMMIT="$GIT_COMMIT" --build-arg BUILD_VERSION="$VERSION" -t automated-approver-unittest:dev -f dockerfiles/DockerfileUnitTest .
  rm -rf vendor
}



dev_local_build() {
  go mod tidy
  go build \
      -tags debug \
      -ldflags "-X 'main.buildDate=$BUILD_DATE' \
                -X 'main.commit=$GIT_COMMIT' \
                -X 'main.version=$VERSION'" \
      -o out/automated-approver \
      gitlab.qredo.com/custody-engine/automated-approver/cmd/service
}


go mod tidy

if [ -n "$1" ]; then
  case $1 in
    docker)
      dev_docker_build
      ;;
    docker_unittest)
      dev_docker_test_build
      ;;
  esac
else
  dev_local_build
fi
