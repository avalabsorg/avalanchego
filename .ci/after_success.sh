#!/usr/bin/env bash

set -ev

bash <(curl -s https://codecov.io/bash)

AVALANCHE_IMAGE=$(docker image ls --format="{{.Repository}}:{{.Tag}}" | grep $DOCKERHUB_REPO | head -n 1)

TRAVIS_TAG="$DOCKERHUB_REPO:travis-$TRAVIS_BUILD_NUMBER"
docker tag $AVALANCHE_IMAGE "$TRAVIS_TAG"

# don't push to dockerhub if this is not being run on the main public repo
# or if it's a PR from a fork ( => secret vars not set )
if [[ $TRAVIS_REPO_SLUG != "ava-labs/avalanchego" || -z "$DOCKER_USERNAME"  ]]; then
  exit 0;
fi

echo "$DOCKER_PASS" | docker login --username "$DOCKER_USERNAME" --password-stdin

# following should push all tags
docker push $DOCKERHUB_REPO 
