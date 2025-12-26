SHELL := /bin/bash
include .env

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