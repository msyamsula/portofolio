start-kubernetes:
	kind create cluster --config kind-config.yaml --name my-cluster

stop-kubernetes:
	kind create cluster --config kind-config.yaml --name my-cluster

load-postgres:
	sh ./binary/postgres/load.sh 

postgres-secret:
	kubectl create secret generic postgres-secret --from-env-file=./binary/postgres/.env

postgres:
	kubectl apply -f ./binary/postgres/deployment.yaml

forward-postgres:
	kubectl port-forward svc/postgres-clusterip 5432:5432

redis-secret:
	kubectl create secret generic redis-secret --from-env-file=./binary/redis/.env

load-redis:
	sh ./binary/redis/load.sh

redis:
	kubectl apply -f ./binary/redis/deployment.yaml

forward-redis:
	kubectl port-forward svc/redis-clusterip 6379:6379

http-secret:
	kubectl create secret generic http-secret --from-env-file=./binary/http/.env

build-http:
	docker build -f ./binary/http/Dockerfile -t syamsuldocker/http:0.0.0 .

load-http:
	kind load docker-image syamsuldocker/http:0.0.0 --name my-cluster

http:
	kubectl apply -f ./binary/http/deployment.yaml

forward-http:
	kubectl port-forward svc/http-clusterip 12000:12000

