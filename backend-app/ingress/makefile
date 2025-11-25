SHELL := /bin/bash
include .env

domain-hosting:
	aws cloudformation deploy --template-file route53.yaml \
	--stack-name hosting \
	--parameter-overrides \
		DomainName=${DomainName} \
		SubDomainGeneral=${SubDomainGeneral} \
		SubDomainUrl=${SubDomainUrl} \
		SubDomainJaeger=${SubDomainJaeger} \
		LoadBalancerDNSName=${LoadBalancerDNSName} \
		LoadBalancerCanonicalHostedZoneId=${LoadBalancerCanonicalHostedZoneId} \
	--capabilities CAPABILITY_NAMED_IAM


destroy-domain-hosting:
	aws cloudformation delete-stack --stack-name hosting
