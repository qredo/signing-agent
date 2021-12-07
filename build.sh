#!/bin/sh

set -e

SERVICE_NAME="core-client"
GIT_COMMIT=$(git rev-list -1 --abbrev-commit HEAD)
BUILD_DATE="$(date -u)"
VERSION="dev"

rm -rf vendor

dev_docker_build() {
  rm -rf vendor
  go mod vendor

  # this fixes go-ethereum build issues caused by Go not vendoring .c files
  # libsecp=`ls -d1 $GOPATH/pkg/mod/github.com/ethereum/go-ethereum*/crypto/secp256k1/libsecp256k1 |tail -1`
  hidapi=`ls -d1 $GOPATH/pkg/mod/github.com/karalabe/usb* | tail -1`
  # cp -rf $libsecp vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/libsecp256k1
  # chmod -R 755 vendor/github.com/ethereum/go-ethereum/crypto/secp256k1/libsecp256k1
  cp -rf $hidapi/* vendor/github.com/karalabe/usb
  chmod -R 755 vendor/github.com/karalabe/usb
  ################################################

  docker build --build-arg BUILD_DATE="$BUILD_DATE" --build-arg BUILD_COMMIT="$GIT_COMMIT" --build-arg BUILD_VERSION="$VERSION" -t "$SERVICE_NAME:dev" -f dockerfiles/Dockerfile .
  rm -rf vendor
}




dev_local_build() {

  go build \
      -tags debug \
      -ldflags "-X 'gitlab.qredo.com/qredo-server/$SERVICE_NAME/service.buildDate=$BUILD_DATE' \
                -X 'gitlab.qredo.com/qredo-server/$SERVICE_NAME/service.commit=$GIT_COMMIT' \
                -X 'gitlab.qredo.com/qredo-server/$SERVICE_NAME/service.version=$VERSION' \
                -X 'gitlab.qredo.com/qredo-server/qredo-core/qerr.InstanceName=$SERVICE_NAME'" \
      -o out/$SERVICE_NAME \
      gitlab.qredo.com/qredo-server/$SERVICE_NAME
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
