#!/bin/bash
#
#  Pushme.sh - this pushs the controller image to a registry that our cluster will pull from when we deploy
#
#   Note: No SemVer here - we do that a bit later for the DeployerCRD
#
# Set these to your registry/repo setup
app="my-operator-controller"
version="latest"
OCI_REG_URL="ghcr.io/dasmlab"

# Tag and push the container up
docker tag $app:scratch ${OCI_REG_URL}/${app}:${version}
docker push ${OCI_REG_URL}/${app}:${version}

