install:
	istioctl install -f install.yaml -y

check:
	kubectl get pods -n istio-system

label:
	kubectl label namespace default istio-injection=enabled

see-label:
	kubectl get namespace --show-labels

restart:
	kubectl rollout restart deployment

verify-sidecar:
	kubectl exec -it url-shortener-deployment-6759bbd787-mpsp5 -c istio-proxy -- curl localhost:15000/config_dump

unlabel:
	kubectl label namespace default istio-injection-

uninstall:
	istioctl uninstall --purge -y


