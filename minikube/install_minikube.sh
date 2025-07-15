#!/bin/bash
#
# A simple minikube setup script.
#
# REQUIRES:  Assumes you have docker installed, kubectl and other deps (doesnt install them for you)
#

set -e

echo "Minikube Dev Environment Installer"

#  Helper function to check for things we need (Docker, Kubectl, minikube binaries, etc.)
function need_cmd() {
  if ! command -v "$1" &>/dev/null; then
    echo "Installing $1..."
    return 1
  else
    echo "$1 already installed."
    return 0
  fi
}

# Check that docker is installed and available. - NOTE: we dont install docker here (assume to be installed)
if ! need_cmd docker; then
  echo "Please install Docker manually and restart this script."
  exit 1
fi

# Check that kubectl is installed and if not, install it
if ! need_cmd kubectl; then
  curl -LO "https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl"
  chmod +x kubectl && sudo mv kubectl /usr/local/bin/
fi

# Check if Minikube binary is on system, if not, install it.
if ! need_cmd minikube; then
  curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
  sudo install minikube-linux-amd64 /usr/local/bin/minikube
  rm minikube-linux-amd64
fi

# Support a --clean option, to clean up and reset Minikube (good for iteration testing).
if [[ "$1" == "--clean" ]]; then
  echo "Cleaning up previous Minikube cluster..."
  minikube delete
fi

# Start minikube using Docker
echo "Starting Minikube using Docker driver..."
minikube start --driver=docker


# Print out Minikube setup info
echo
echo "*** Minikube Info ***"
minikube status
kubectl version --client=true
echo "Kubernetes nodes:"
kubectl get nodes

echo
echo "Minikube setup and running."

