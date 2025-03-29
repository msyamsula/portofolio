build:
	docker build -t main-page .

run:
	make build
	docker-compose up -d

down:
	docker-compose down

push:
	make build
	docker tag main-page syamsuldocker/main-page:0.0.3
	docker push syamsuldocker/main-page:0.0.3

deploy:
	make down
	make push
	kubectl --kubeconfig=../../kubeconfig.yaml apply -f deployment.yaml
