#!/bin/bash
#
# cycleme.sh - a helper script to cycle your controller deployments to a cluster for development iteration
#
#  Assumes you have all the names, vars and such aligned with the other scripts (buildme, pushme) and your credentials, registry uri / repo
make undeploy uninstall
./buildme.sh
./pushme.sh 
make install deploy
./create-registry-secret.sh
