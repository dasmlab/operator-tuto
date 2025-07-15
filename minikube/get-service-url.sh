# Get Service URL (Minikube)
#
#  This will open a browser and auto-port-forward (if you have the right setup) for a Docker minikube service for dialogue-app
#  This tries to open a browser, so I assume you have X or a local browser ability on your terminal, etc.

minikube service -n dialogue-ns dialogue-app 
