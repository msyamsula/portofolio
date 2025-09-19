start-kubernetes:
	kind create cluster --config kind-config.yaml --name my-cluster

stop-kubernetes:
	kind delete cluster --name my-cluster

load-postgres:
	sh ./binary/postgres/load.sh 

postgres-secret:
	kubectl create secret generic postgres-secret --from-env-file=./binary/postgres/.env --dry-run=client -o yaml | kubectl apply -f -

postgres:
	kubectl apply -f ./binary/postgres/deployment.yaml

forward-postgres:
	kubectl port-forward svc/postgres-clusterip 5432:5432

redis-secret:
	kubectl create secret generic redis-secret --from-env-file=./binary/redis/.env --dry-run=client -o yaml | kubectl apply -f -

load-redis:
	sh ./binary/redis/load.sh

redis:
	kubectl apply -f ./binary/redis/deployment.yaml

forward-redis:
	kubectl port-forward svc/redis-clusterip 6379:6379

http-secret:
	kubectl create secret generic http-secret --from-env-file=./binary/http/.env --dry-run=client -o yaml | kubectl apply -f -

build-http:
	docker build -f ./binary/http/Dockerfile -t syamsuldocker/http:0.0.0 .

load-http:
	kind load docker-image syamsuldocker/http:0.0.0 --name my-cluster

http:
	kubectl apply -f ./binary/http/deployment.yaml

forward-http:
	kubectl port-forward svc/http-clusterip 12000:12000

build-main-page:
	docker build -f ./ui/main-page/Dockerfile -t syamsuldocker/main-page:0.0.0 ./ui/main-page

load-main-page:
	kind load docker-image syamsuldocker/main-page:0.0.0 --name my-cluster

run-main-page:
	kubectl apply -f ./ui/main-page/deployment.yaml

forward-main-page:
	kubectl port-forward svc/main-page-clusterip 13000:80

delete-main-page:
	kubectl delete deployment main-page-deployment

prepare-main-page: delete-main-page build-main-page load-main-page run-main-page

