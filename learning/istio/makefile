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

verify:
	kubectl exec -it url-shortener-deployment-5cb7dc8d49-q775g -c istio-proxy -- curl localhost:15000/config_dump

unlabel:
	kubectl label namespace default istio-injection-

uninstall:
	istioctl uninstall --purge -y


