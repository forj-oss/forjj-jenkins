#!/usr/bin/env bash

set -e

if [ "$BUILD_ENV_LOADED" != "true" ]
then
   echo "Please go to your project and load your build environment. 'source build-env.sh'"
   exit 1
fi

TAG="-t $(awk '$0 ~ /^ *image: ".*"$/ { print $0 }' jenkins.yaml | sed 's/^ *image: "*\(.*\)".*$/\1/g')"

cd $BUILD_ENV_PROJECT

if [ "$http_proxy" != "" ]
then
   PROXY="--build-arg http_proxy=$http_proxy --build-arg https_proxy=$http_proxy --build-arg no_proxy=$no_proxy"
fi

create-go-build-env.sh

glide i

go build

set -x
$BUILD_ENV_DOCKER build $PROXY $DOCKERFILE $TAG .
