#!/bin/sh
set -eu

CONFIG_FILE=/opt/kafka/config/server.properties

/opt/kafka/bin/kafka-storage.sh format --ignore-formatted --standalone -t "${KAFKA_KRAFT_CLUSTER_ID}" -c "$CONFIG_FILE"
exec /opt/kafka/bin/kafka-server-start.sh "$CONFIG_FILE"
