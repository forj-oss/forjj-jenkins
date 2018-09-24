#!/bin/bash
#

if [[ "$DOCKER_DOOD_GROUP" != "" ]]
then
    echo "Configuring docker socket group and user devops"
    addgroup -g $DOCKER_DOOD_GROUP docker 
    addgroup devops docker
fi