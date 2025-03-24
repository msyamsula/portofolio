run:
	docker-compose -f tech-stack/nsq/nsqlookupd/docker-compose.yaml up -d
	docker-compose -f tech-stack/postgres/docker-compose.yaml up -d
	docker-compose -f tech-stack/redis/docker-compose.yaml up -d
	
	docker-compose -f ui/chat/docker-compose.yaml up -d
	docker-compose -f ui/graph/docker-compose.yaml up -d
	docker-compose -f ui/main-page/docker-compose.yaml up -d
	docker-compose -f ui/url/docker-compose.yaml up -d
	
	docker-compose -f tech-stack/nsq/nsqadmin/docker-compose.yaml up -d
	docker-compose -f tech-stack/nsq/nsqd/docker-compose.yaml up -d
	docker-compose -f tech-stack/nsq/consumer/docker-compose.yaml up -d
	docker-compose -f tech-stack/telemetry/jaeger/docker-compose.yaml up -d
	docker-compose -f tech-stack/telemetry/prometheus/docker-compose.yaml up -d

	docker-compose -f tech-stack/websocket/docker-compose.yaml up -d
	docker-compose -f tech-stack/http/docker-compose.yaml up -d



down:
	docker-compose -f tech-stack/nsq/nsqlookupd/docker-compose.yaml down
	docker-compose -f tech-stack/postgres/docker-compose.yaml down
	docker-compose -f tech-stack/redis/docker-compose.yaml down

	docker-compose -f tech-stack/telemetry/jaeger/docker-compose.yaml down
	docker-compose -f tech-stack/telemetry/prometheus/docker-compose.yaml down
	docker-compose -f tech-stack/http/docker-compose.yaml down
	docker-compose -f tech-stack/nsq/nsqadmin/docker-compose.yaml down
	docker-compose -f tech-stack/nsq/nsqd/docker-compose.yaml down
	docker-compose -f tech-stack/nsq/consumer/docker-compose.yaml down
	docker-compose -f tech-stack/websocket/docker-compose.yaml down

	docker-compose -f ui/chat/docker-compose.yaml down
	docker-compose -f ui/graph/docker-compose.yaml down
	docker-compose -f ui/main-page/docker-compose.yaml down
	docker-compose -f ui/url/docker-compose.yaml down