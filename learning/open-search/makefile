start:
	docker run -itd --net test --name opensearch \
	-p 9200:9200 -p 9600:9600 \
	-e "discovery.type=single-node" -e "DISABLE_SECURITY_PLUGIN=true" \
	opensearchproject/opensearch:latest

dashboard:
	docker run -itd --name dashboard \
   --network test \
   -p 5601:5601 \
   -e "DISABLE_SECURITY_DASHBOARDS_PLUGIN=true" \
   -v ./opensearch_dashboards.yml:/usr/share/opensearch-dashboards/config/opensearch_dashboards.yml \
   opensearchproject/opensearch-dashboards:latest