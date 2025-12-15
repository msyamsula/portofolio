SHELL := /bin/bash
include .env

# start the cluster
start:
	kind create cluster --config kind-config.yaml --name ${CLUSTER}
	docker network connect test ${CLUSTER}-control-plane
	docker network connect test ${CLUSTER}-worker
	docker network connect test ${CLUSTER}-worker2
	docker network connect test ${CLUSTER}-worker3

stop:
	kind delete cluster --name ${CLUSTER}

# namespace
obs:
	kubectl config set-context --current --namespace=obs

istio-system:
	kubectl config set-context --current --namespace=istio-system

default:
	kubectl config set-context --current --namespace=default

elastic:
	kubectl config set-context --current --namespace=elastic-system


# context management, if you have multiple kubernetes
list:
	kubectl config get-contexts

use:
	kubectl config use-context kind-${CLUSTER}

peek:
	kubectl config current-context

rename:
	kubectl config rename-context ${CONTEXT} ${NEW_CONTEXT}

remove:
	kubectl config delete-context ${CONTEXT}

forward:
	kubectl port-forward svc/${APP} ${PORT}:${PORT} -n ${NAMESPACE}


## loud balancer provider in kind
## in cloud provider you do not have to do this
## this install, run, and allow is just for mimicking cloud behavior in local
install-lb-cloud-provider:
	brew install cloud-provider-kind

run-lb-provider:
	sudo cloud-provider-kind


# By default, Kubernetes expects workloads will not run on control plane nodes and 
# labels them with node.kubernetes.io/exclude-from-external-load-balancers, 
# which stops load balancers from accessing them.
# If you are running workloads on control plane nodes, as is the default kind configuration, 
# you will need to remove this label to access them using a LoadBalancer:
allow-lb-control-plane:
	kubectl label node ${CLUSTER}-control-plane node.kubernetes.io/exclude-from-external-load-balancers-


# istio
install:
	istioctl install -f install.yaml -y

check:
	kubectl get pods -n istio-system

label:
	kubectl label namespace default istio-injection=enabled

label-obs:
	kubectl label namespace obs istio-injection=enabled

see-label:
	kubectl get namespace --show-labels

restart:
	kubectl rollout restart deployment

verify:
	kubectl exec -it user-74674c96-7pmsd -c istio-proxy -- curl localhost:15000/config_dump

unlabel:
	kubectl label namespace default istio-injection-

unlabel-obs:
	kubectl label namespace obs istio-injection-

uninstall:
	istioctl uninstall --purge -y



restart-app:
	kubectl rollout restart deploy



# elastic search
load-elastic:
	kind load docker-image docker.elastic.co/elasticsearch/elasticsearch:9.2.1 --name ${CLUSTER}
	






