SHELL := /bin/bash
include .env

# jaeger
jaeger-setup-env:
	kubectl create configmap jaeger-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f - -n obs

jaeger:
	kubectl apply -f jaeger.yaml -n obs

delete-jaeger:
	kubectl delete -f jaeger.yaml -n obs


# prometheus
prometheus:
	kubectl apply -f prom.yaml -n obs

# kiali
kiali:
	kubectl apply -f kiali.yaml -n obs

# grafana
grafana:
	kubectl apply -f grafana.yaml -n obs

# fluentd
fluentd:
	kubectl apply -f fluentd.yaml -n obs

fluentd-delete:
	kubectl delete -f fluentd.yaml -n obs

fluentd-setup-env:
	kubectl create configmap fluentd-env --from-env-file=.env --dry-run=client -o yaml | kubectl apply -f - -n obs