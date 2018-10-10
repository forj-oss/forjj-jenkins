#!/bin/bash
#
# This script is used to publish officially all released docker images (tagged)
#
# Release workflow is:
#
# - Someone fork and create a tag release then submit a PR.
# - GitHub jenkins can be started to start an 'ITG' image validation
# - The repo maintainer at some time will accept the new release.
# - Github should send a jenkins job to build officially this new release
#   I expect to get this info in $1 (Release number)

# Then this job should implement the following code in jenkins
# And jenkins-ci images for each flavors will be officially pushed to the internal registry.

if [ "$BUILD_ENV_LOADED" != "true" ]
then
   echo "Please go to your project and load your build environment. 'source build-env.sh'"
   exit 1
fi

if [[ "$CI_ENABLED" != "FALSE" ]]
then
  if [ "$GITHUB_TOKEN" = '' ]
  then
    echo "GITHUB_TOKEN is missing. You need it to publish."
    exit 1
  fi

  if [[ "$DOCKERHUB_PASSWORD" = '' ]] || [[ "$DOCKERHUB_USERNAME" = '' ]]
  then
    echo "DOCKERHUB_PASSWORD and/or DOCKERHUB_USERNAME is missing. You need it to publish."
    exit 1
  fi

  set -e
  docker login -u $DOCKERHUB_USERNAME -p $DOCKERHUB_PASSWORD
  set +e
fi

TAG_BASE="$(eval "echo $(awk '$1 ~ /image:/ { print $2 }' jenkins.yaml)")"

if [ ! -f releases.lst ]
then
   echo "VERSION or releases.lst files not found. Please move to the repo root dir and call back this script."
   exit 1
fi

if [[ "$CI_ENABLED" = "FALSE" ]]
then
  echo  "You are going to publish manually. This is not recommended. 
Do it only if Jenkins fails to do it automatically.

Press Enter to continue."
  read

  if [ "$(git rev-parse --abbrev-ref HEAD)" != master ]
  then
    echo "You must be on master branch."
    exit 1
  fi
  REMOTE="$(git remote -v | grep "^upstream")"
  if [ "$REMOTE" = "" ]
  then
      echo "upstream is missing. You must have it configured (git@github.com:forj-oss/forjj-jenkins.git) and rights given to push"
      exit 1
  fi
  if [[ ! "$REMOTE" =~ git@github\.com:forj-oss/forjj-jenkins\.git ]]
  then
      echo "upstream is wrongly configured. It must be set with git@github.com:forj-oss/forjj-jenkins.git"
      exit 1
  fi
  git stash # Just in case
  git fetch upstream
  git reset --hard upstream/master
fi

case "$1" in
  release-it )
    VERSION=$(eval "echo $(awk '$1 ~ /version:/ { print $2 }' jenkins.yaml)")
    if [ "$(git tag -l $VERSION)" = "" ]
    then
       echo "Unable to publish a release version. git tag missing"
       exit 1
    fi
    COMMIT="$(git log -1 --oneline| cut -d ' ' -f 1)"
    if [ "$(git tag -l --points-at $COMMIT | grep $VERSION)" = "" ]
    then
       echo "'$COMMIT' is not tagged with '$VERSION'. Only commit tagged can publish officially this tag as docker image."
       exit 1
    fi
    VERSION_TAG=${VERSION}_

    TAG="$(grep VERSION version.go | sed 's/const VERSION="\(.*\)"/\1/g')"
    PRE_RELEASE="$(grep VERSION version.go | sed 's/const PRERELEASE="\(.*\)"/\1/g')"
    if [ "$(git tag | grep "^$TAG$")" != "" ]
    then
      echo "Unable to publish $TAG. Already published and released."
      exit 1
    fi
    if [[ "$1" != "--auto" ]] && [[ "$CI_ENABLED" = "FALSE" ]]
    then
      echo "You are going to publish version $TAG. Ctrl-C to interrupt or press Enter to go on"
      read
    else
      echo "Publishing version $TAG..."
    fi
    ;;
  latest )
    VERSION=latest
    VERSION_TAG=latest_
    TAG=latest
    git fetch --tags
    git tag -d $TAG
    ;;
  *)
    echo "Script used to publish release and latest code ONLY. If you want to test a fork, use build. It will create a local docker image $TAG_BASE:test"
    exit 1
esac

echo "Tagging to $TAG..."
git tag $TAG

echo "Pushing it ..."
if [[ "$CI_ENABLED" = "TRUE" ]]
then
    git config --local credential.helper 'store --file /tmp/.git.store'
    echo "https://${GITHUB_USER}:${GITHUB_TOKEN}@github.com" > /tmp/.git.store
    git push -f origin $TAG
    rm -f /tmp/.git.store
    GOPATH=go-workspace
else
    git push -f upstream $TAG

    build.sh
fi

echo "Deploying $BE_PROJECT to dockerhub..."

cat releases.lst | while read LINE
do
   [[ "$LINE" =~ ^# ]] && continue
   TAGS="$(echo "$LINE" | awk -F'|' '{ print $2 }' | sed 's/,/ /g')"
   echo "=============== Building devops/$(basename $BUILD_ENV_PROJECT)"
   $(dirname $0)/build.sh
   echo "=============== Publishing tags"
   for TAG in $TAGS
   do
      echo "=> $TAG_BASE:$TAG"
      docker tag $TAG_BASE $TAG_BASE:$TAG
      docker push $TAG_BASE:$TAG
   done
   echo "=============== DONE"
done
