#!/bin/bash
#
#  Buildme.sh - builds a instance of the my-operator-controller:scratch (locally)
app="my-operator-controller"
version="scratch"
docker build -t $app:$version .

