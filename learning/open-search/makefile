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


elastic-start:
	docker run -itd --name elastic --net test -p 9200:9200 -it -m 1GB docker.elastic.co/elasticsearch/elasticsearch:9.2.1

elastic-password:
	docker exec -it ${ELASTIC_HOST} /usr/share/elasticsearch/bin/elasticsearch-reset-password -u elastic

elastic-token:
	docker exec -it ${ELASTIC_HOST} /usr/share/elasticsearch/bin/elasticsearch-create-enrollment-token -s kibana

elastic-cert:
	docker cp ${ELASTIC_HOST}:/usr/share/elasticsearch/config/certs/http_ca.crt .

elastic-test:
	curl --cacert http_ca.crt -u elastic:${ELASTICSEARCH_PASSWORD} https://localhost:9200/_cat/indices?v

elastic-save-cert:
	kubectl create secret generic es-cert --from-file=http_ca.crt=http_ca.crt -n obs

# kibana
kibana-start:
	docker run -itd --name kibana --net test -p 5601:5601 docker.elastic.co/kibana/kibana:9.2.1