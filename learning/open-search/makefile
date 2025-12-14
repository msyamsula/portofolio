SHELL := /bin/bash
include .env

start:
	docker run -itd -p 9200:9200 -p 9600:9600 \
	--name opensearch \
	--net test \
	-e "discovery.type=single-node" -e \
	"OPENSEARCH_INITIAL_ADMIN_PASSWORD=${OPENSEARCH_PASSWORD}" \
	opensearchproject/opensearch:latest


dashboard:
	docker run -itd --name dashboard \
   --network test \
   -p 5601:5601 \
   -v ./opensearch_dashboards.yml:/usr/share/opensearch-dashboards/config/opensearch_dashboards.yml \
   opensearchproject/opensearch-dashboards:latest