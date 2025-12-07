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



# url-shortener
url-shortener-setup-env:
	kubectl create configmap url-shortener-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f - -n default

url-shortener-run:
	kubectl apply -f url-shortener.yaml -n default

url-shortener-deploy:
	kubectl set image deployment/url-shortener url-shortener=docker.io/library/url-shortener:${VERSION} -n default

url-shortener-delete:
	kubectl delete -f url-shortener.yaml -n default

url-shortener-load:
	kind load docker-image url-shortener:${VERSION} --name ${CLUSTER}

# user
user-setup-env:
	kubectl create configmap user-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f - -n default

user-run:
	kubectl apply -f user.yaml -n default

user-deploy:
	kubectl set image deployment/user user=docker.io/library/user:${VERSION} -n default

user-delete:
	kubectl delete -f user.yaml -n default

user-load:
	kind load docker-image user:${VERSION} --name ${CLUSTER}


restart-app:
	kubectl rollout restart deploy


ingress:
	kubectl apply -f domain-gateway.yaml -n default


## observability namespace
jaeger:
	kubectl apply -f jaeger.yaml -n obs

delete-jaeger:
	kubectl delete -f jaeger.yaml -n obs

prometheus:
	kubectl apply -f prom.yaml -n obs

prometheus-restart:
	kubectl rollout restart deployment prometheus -n obs

# kiali
kiali:
	kubectl apply -f kiali.yaml -n obs

kiali-ui:
	istioctl dashboard kiali -n obs

delete-kiali:
	kubectl delete -f kiali.yaml -n obs 

# grafana
grafana:
	kubectl apply -f grafana.yaml -n obs

destroy-grafana:
	kubectl delete -f grafana.yaml -n obs

# elastic search
load-elastic:
	kind load docker-image docker.elastic.co/elasticsearch/elasticsearch:9.2.1 --name ${CLUSTER}

elastic-start:
	docker run -itd --name elastic --net test -p 9200:9200 -it -m 1GB docker.elastic.co/elasticsearch/elasticsearch:9.2.1

elastic-password:
	docker exec -it ${ELASTIC_HOST} /usr/share/elasticsearch/bin/elasticsearch-reset-password -u elastic

elastic-token:
	docker exec -it ${ELASTIC_HOST} /usr/share/elasticsearch/bin/elasticsearch-create-enrollment-token -s kibana

elastic-cert:
	docker cp ${ELASTIC_HOST}:/usr/share/elasticsearch/config/certs/http_ca.crt .

elastic-test:
	curl --cacert http_ca.crt -u elastic:${ELASTIC_PASSWORD} https://localhost:9200/_cat/indices?v

elastic-save-cert:
	kubectl create secret generic es-cert --from-file=http_ca.crt=http_ca.crt -n obs

# kibana
kibana-start:
	docker run -itd --name kibana --net test -p 5601:5601 docker.elastic.co/kibana/kibana:9.2.1

# fluentd
fluentd:
	kubectl apply -f fluentd.yaml -n obs

fluentd-delete:
	kubectl delete -f fluentd.yaml -n obs

fluentd-setup-env:
	kubectl create configmap fluentd-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f - -n obs


# hit
hit:
	watch -n 1 curl -o /dev/null -s -w http://172.19.0.4/short?long_url=https://mail.google.com/mail/u/0/%23search/nameserver/FMfcgzQcqtkJtqcMJDxDmQpgbZMbMksz





