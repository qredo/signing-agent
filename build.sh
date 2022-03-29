#!/bin/sh

set -e

GIT_COMMIT=$(git rev-list -1 --abbrev-commit HEAD)
BUILD_DATE="$(date -u)"
VERSION="dev"

rm -rf vendor

dev_docker_build() {

  docker build --build-arg BUILD_DATE="$BUILD_DATE" --build-arg BUILD_COMMIT="$GIT_COMMIT" --build-arg BUILD_VERSION="$VERSION" -t "$SERVICE_NAME:dev" -f dockerfiles/Dockerfile .
  rm -rf vendor
}



dev_local_build() {

  go build \
      -tags debug \
      -ldflags "-X 'main.buildDate=$BUILD_DATE' \
                -X 'main.commit=$GIT_COMMIT' \
                -X 'main.version=$VERSION'" \
      -o out/core-client \
      gitlab.qredo.com/qredo-server/core-client/cmd/service
}


go mod tidy

if [ -n "$1" ]; then
  case $1 in
    docker)
      dev_docker_build
      ;;
  esac
else
  dev_local_build
fi
