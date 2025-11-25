SHELL := /bin/bash
include .env

# start the cluster
start:
	kind create cluster --config kind-config.yaml --name ${CLUSTER}
	docker network connect test ${CLUSTER}-control-plane
	docker network connect test ${CLUSTER}-worker

stop:
	kind delete cluster --name ${CLUSTER}

# context management, if you have multiple kubernetes
list:
	kubectl config get-contexts

use:
	kubectl config use-context ${CONTEXT}

peek:
	kubectl config current-context

rename:
	kubectl config rename-context ${CONTEXT} ${NEW_CONTEXT}

remove:
	kubectl config delete-context ${CONTEXT}

forward:
	kubectl port-forward svc/${APP}-clusterip ${PORT}:${PORT} -n ${NAMESPACE}


# istio
install:
	istioctl install -f install.yaml -y

check:
	kubectl get pods -n istio-system

label:
	kubectl label namespace default istio-injection=enabled

see-label:
	kubectl get namespace --show-labels

restart:
	kubectl rollout restart deployment

verify:
	kubectl exec -it url-shortener-deployment-5cb7dc8d49-q775g -c istio-proxy -- curl localhost:15000/config_dump

unlabel:
	kubectl label namespace default istio-injection-

uninstall:
	istioctl uninstall --purge -y



# deploy app
run:
	kubectl create configmap ${APP}-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f deployment.yaml

deploy:
	kubectl set image deployment/${APP}-deployment ${APP}=docker.io/library/${APP}:${VERSION}

delete:
	kubectl delete -f deployment.yaml

load:
	kind load docker-image ${APP}:${VERSION} --name ${CLUSTER}

restart-app:
	kubectl rollout restart deploy


ingress:
	kubectl apply -f domain-gateway.yaml


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


## observability namespace
jaeger:
	kubectl apply -f jaeger.yaml

delete-jaeger:
	kubectl delete -f jaeger.yaml

prometheus:
	kubectl apply -f prom.yaml 

prometheus-restart:
	kubectl rollout restart deployment prometheus 

# kiali
kiali:
	kubectl apply -f kiali.yaml 

delete-kiali:
	kubectl apply -f kiali.yaml 

kiali-ui:
	istioctl dashboard kiali 

delete-kiali:
	kubectl delete -f kiali.yaml 


# hit
hit:
	watch -n 1 curl -o /dev/null -s -w http://172.19.0.4/short?long_url=https://mail.google.com/mail/u/0/%23search/nameserver/FMfcgzQcqtkJtqcMJDxDmQpgbZMbMksz





