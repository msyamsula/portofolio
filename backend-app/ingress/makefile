SHELL := /bin/bash
include .env

deploy-hosting:
	aws cloudformation deploy --template-file route53.yaml \
	--stack-name hosting \
	--parameter-overrides \
		DomainName=${DomainName} \
		SubDomainGeneral=${SubDomainGeneral} \
		SubDomainUrl=${SubDomainUrl} \
		SubDomainJaeger=${SubDomainJaeger} \
		SubDomainPrometheus=${SubDomainPrometheus} \
		LoadBalancerDNSName=${LoadBalancerDNSName} \
		LoadBalancerCanonicalHostedZoneId=${LoadBalancerCanonicalHostedZoneId} \
	--capabilities CAPABILITY_NAMED_IAM


destroy-hosting:
	aws cloudformation delete-stack --stack-name hosting

ingress:
	kubectl apply -f domain-gateway.yaml
